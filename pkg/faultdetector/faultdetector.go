// Package faultdetector implements Optimism fault detector.
package faultdetector

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/LiskHQ/op-fault-detector/pkg/chain"
	"github.com/LiskHQ/op-fault-detector/pkg/config"
	"github.com/LiskHQ/op-fault-detector/pkg/encoding"
	"github.com/LiskHQ/op-fault-detector/pkg/log"
	"github.com/LiskHQ/op-fault-detector/pkg/utils/notification"
	"github.com/ethereum/go-ethereum/common"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	serviceIntervalInSeconds = 1
	waitTimeInFailure        = 10
)

// FaultDetector contains all the RPC providers/contract accessors and holds state information.
type FaultDetector struct {
	ctx                    context.Context
	logger                 log.Logger
	errorChan              chan error
	wg                     *sync.WaitGroup
	metrics                *faultDetectorMetrics
	l1RpcApi               *chain.ChainAPIClient
	l2RpcApi               *chain.ChainAPIClient
	oracleContractAccessor OracleAccessor
	faultProofWindow       uint64
	currentOutputIndex     uint64
	diverged               bool
	ticker                 *time.Ticker
	quitTickerChan         chan struct{}
	notification           *notification.Notification
	mutex                  *sync.RWMutex
}

type faultDetectorMetrics struct {
	highestOutputIndex   prometheus.Gauge
	stateMismatch        prometheus.Gauge
	apiConnectionFailure prometheus.Gauge
}

// NewFaultDetectorMetrics returns [FaultDetectorMetrics] with initialized metrics and registering to prometheus registry.
func NewFaultDetectorMetrics(reg prometheus.Registerer) *faultDetectorMetrics {
	m := &faultDetectorMetrics{
		highestOutputIndex: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "fault_detector_highest_output_index",
				Help: "The highest current output index that is being checked for faults",
			}),
		stateMismatch: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "fault_detector_is_state_mismatch",
			Help: "0 when state is matched, 1 when mismatch",
		}),
		apiConnectionFailure: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "fault_detector_api_connection_failure",
			Help: "Number of times API call failed",
		}),
	}
	reg.MustRegister(m.highestOutputIndex)
	reg.MustRegister(m.stateMismatch)
	reg.MustRegister(m.apiConnectionFailure)

	return m
}

// NewFaultDetector will return [FaultDetector] with the initialized providers and configuration.
func NewFaultDetector(ctx context.Context, logger log.Logger, errorChan chan error, wg *sync.WaitGroup, faultDetectorConfig *config.FaultDetectorConfig, metricRegistry *prometheus.Registry, notification *notification.Notification) (*FaultDetector, error) {
	// Initialize API Providers
	l1RpcApi, err := chain.GetAPIClient(ctx, faultDetectorConfig.L1RPCEndpoint, logger)
	if err != nil {
		logger.Errorf("Failed to create API client for L1 Provider with given endpoint: %s, error: %v", faultDetectorConfig.L1RPCEndpoint, err)
		return nil, err
	}

	l2RpcApi, err := chain.GetAPIClient(ctx, faultDetectorConfig.L2RPCEndpoint, logger)
	if err != nil {
		logger.Errorf("Failed to create API client for L2 Provider with given endpoint: %s, error: %v", faultDetectorConfig.L2RPCEndpoint, err)
		return nil, err
	}

	l2ChainID, err := l2RpcApi.GetChainID(ctx)
	if err != nil {
		logger.Errorf("Failed to get L2 provider's chainID: %d, error: %v", encoding.MustConvertBigIntToUint64(l2ChainID), err)
		return nil, err
	}

	// Initialize Oracle contract accessor
	chainConfig := &chain.ConfigOptions{
		L1RPCEndpoint:                 faultDetectorConfig.L1RPCEndpoint,
		ChainID:                       encoding.MustConvertBigIntToUint64(l2ChainID),
		L2OutputOracleContractAddress: faultDetectorConfig.L2OutputOracleContractAddress,
	}

	oracleContractAccessor, err := chain.NewOracleAccessor(ctx, chainConfig)
	if err != nil {
		logger.Errorf("Failed to create Oracle contract accessor with chainID: %d, L1 endpoint: %s and L2OutputOracleContractAddress: %s, error: %v", encoding.MustConvertBigIntToUint64(l2ChainID), faultDetectorConfig.L1RPCEndpoint, faultDetectorConfig.L2OutputOracleContractAddress, err)
		return nil, err
	}

	finalizedPeriodSeconds, err := oracleContractAccessor.FinalizationPeriodSeconds()
	if err != nil {
		logger.Errorf("Failed to query `FinalizationPeriodSeconds` from Oracle contract accessor, error: %v", err)
		return nil, err
	}

	logger.Infof("Fault proof window is set to %d.", finalizedPeriodSeconds)

	var currentOutputIndex uint64
	if faultDetectorConfig.StartBatchIndex == -1 {
		logger.Infof("Finding appropriate starting unfinalized batch....")
		firstUnfinalized, _ := FindFirstUnfinalizedOutputIndex(
			ctx,
			logger,
			encoding.MustConvertBigIntToUint64(finalizedPeriodSeconds),
			oracleContractAccessor,
			l2RpcApi,
		)
		if firstUnfinalized == 0 {
			logger.Infof("No unfinalized batches found. skipping all batches.")
			nextOutputIndex, err := oracleContractAccessor.GetNextOutputIndex()
			if err != nil {
				logger.Errorf("Failed to query next output index, error: %v", err)
				return nil, err
			}
			currentOutputIndex = encoding.MustConvertBigIntToUint64(nextOutputIndex) - 1
		} else {
			currentOutputIndex = firstUnfinalized
		}
	} else {
		currentOutputIndex = uint64(faultDetectorConfig.StartBatchIndex)
	}
	logger.Infof("Starting unfinalized batch index is set to %d.", currentOutputIndex)

	metrics := NewFaultDetectorMetrics(metricRegistry)
	// Initially set state mismatch to 0
	metrics.stateMismatch.Set(0)

	faultDetector := &FaultDetector{
		ctx:                    ctx,
		logger:                 logger,
		errorChan:              errorChan,
		wg:                     wg,
		l1RpcApi:               l1RpcApi,
		l2RpcApi:               l2RpcApi,
		oracleContractAccessor: oracleContractAccessor,
		faultProofWindow:       finalizedPeriodSeconds.Uint64(),
		currentOutputIndex:     currentOutputIndex,
		diverged:               false,
		metrics:                metrics,
		notification:           notification,
		mutex:                  new(sync.RWMutex),
	}

	return faultDetector, nil
}

// Start will start the fault detector service by invoking the service every given interval.
func (fd *FaultDetector) Start() {
	defer fd.wg.Done()
	fd.logger.Infof("Started fault detector service, checking for state root every %d seconds.", serviceIntervalInSeconds)
	fd.ticker = time.NewTicker(serviceIntervalInSeconds * time.Second)
	fd.quitTickerChan = make(chan struct{})
	for {
		select {
		case <-fd.ticker.C:
			if err := fd.checkFault(); err != nil {
				time.Sleep(waitTimeInFailure * time.Second)
			}
		case <-fd.quitTickerChan:
			fd.logger.Infof("Quit ticker for periodic fault detection.")
			return
		}
	}
}

// Stop will stop the ticker.
func (fd *FaultDetector) Stop() {
	fd.ticker.Stop()
	close(fd.quitTickerChan)
	fd.logger.Infof("Successfully stopped fault detector service.")
}

// checkFault continuously checks for the faults at regular interval.
func (fd *FaultDetector) checkFault() error {
	startTime := time.Now()
	fd.logger.Infof("Checking current batch with output index: %d.", fd.currentOutputIndex)

	nextOutputIndex, err := fd.oracleContractAccessor.GetNextOutputIndex()
	if err != nil {
		fd.logger.Errorf("Failed to query next output index, error: %v.", err)
		fd.metrics.apiConnectionFailure.Inc()
		return err
	}

	latestBatchIndex := encoding.MustConvertBigIntToUint64(nextOutputIndex) - 1
	fd.logger.Infof("Latest batch index is set to %d.", latestBatchIndex)
	if fd.currentOutputIndex > latestBatchIndex {
		fd.logger.Infof("Current output index %d is ahead of the oracle latest batch index %d. Waiting...", fd.currentOutputIndex, latestBatchIndex)
		return fmt.Errorf("current output index is ahead of the oracle latest batch index")
	}

	l2OutputData, err := fd.oracleContractAccessor.GetL2Output(encoding.MustConvertUint64ToBigInt(fd.currentOutputIndex))
	if err != nil {
		fd.logger.Errorf("Failed to fetch output associated with index: %d, error: %v.", fd.currentOutputIndex, err)
		fd.metrics.apiConnectionFailure.Inc()
		return err
	}

	latestBlockNumber, err := fd.l2RpcApi.GetLatestBlockNumber(fd.ctx)
	if err != nil {
		fd.logger.Errorf("Failed to query L2 latest block number: %d, error: %v", latestBlockNumber, err)
		fd.metrics.apiConnectionFailure.Inc()
		return err
	}

	l2OutputBlockNumber := l2OutputData.L2BlockNumber
	expectedOutputRoot := l2OutputData.OutputRoot
	if latestBlockNumber < l2OutputBlockNumber {
		fd.logger.Infof("L2 node is behind, waiting for node to sync with the network...")
		return fmt.Errorf("l2 node is behind")
	}

	outputBlockHeader, err := fd.l2RpcApi.GetBlockHeaderByNumber(fd.ctx, encoding.MustConvertUint64ToBigInt(l2OutputBlockNumber))
	if err != nil {
		fd.logger.Errorf("Failed to fetch block header by number: %d, error: %v.", l2OutputBlockNumber, err)
		fd.metrics.apiConnectionFailure.Inc()
		return err
	}

	messagePasserProofResponse, err := fd.l2RpcApi.GetProof(fd.ctx, encoding.MustConvertUint64ToBigInt(l2OutputBlockNumber), common.HexToAddress(chain.L2BedrockMessagePasserAddress))
	if err != nil {
		fd.logger.Errorf("Failed to fetch message passer proof for the block with height: %d and address: %s, error: %v.", l2OutputBlockNumber, chain.L2BedrockMessagePasserAddress, err)
		fd.metrics.apiConnectionFailure.Inc()
		return err
	}

	calculatedOutputRoot := encoding.ComputeL2OutputRoot(
		outputBlockHeader.Root,
		messagePasserProofResponse.StorageHash,
		outputBlockHeader.Hash(),
	)
	if calculatedOutputRoot != expectedOutputRoot {
		fd.mutex.Lock()
		fd.diverged = true
		fd.mutex.Unlock()

		fd.metrics.stateMismatch.Set(1)
		finalizationTime := time.Unix(int64(outputBlockHeader.Time+fd.faultProofWindow), 0)

		if fd.notification != nil {
			notificationMessage := fmt.Sprintf("*Fault detected*, state root does not match:\noutputIndex: %d\nExpectedStateRoot: %s\nCalculatedStateRoot: %s\nFinalizationTime: %s", fd.currentOutputIndex, expectedOutputRoot, calculatedOutputRoot, finalizationTime)
			if err := fd.notification.Notify(notificationMessage); err != nil {
				fd.logger.Errorf("Error while sending notification, %v", err)
			}
		}

		fd.logger.Errorf("State root does not match expectedStateRoot: %s, calculatedStateRoot: %s, finalizationTime: %s.", expectedOutputRoot, calculatedOutputRoot, finalizationTime)
		return nil
	}

	fd.metrics.highestOutputIndex.Set(float64(fd.currentOutputIndex))

	// Time taken to execute each batch in milliseconds.
	elapsedTime := time.Since(startTime).Milliseconds()
	fd.logger.Infof("Successfully checked current batch with index %d --> ok, time taken %dms.", fd.currentOutputIndex, elapsedTime)
	fd.mutex.Lock()
	fd.diverged = false
	fd.mutex.Unlock()

	fd.currentOutputIndex++
	fd.metrics.stateMismatch.Set(0)
	return nil
}

// IsFaultDetected returns status of the fault detector.
func (fd *FaultDetector) IsFaultDetected() bool {
	fd.mutex.RLock()
	defer fd.mutex.RUnlock()
	return fd.diverged
}

// GetFaultDetector create [FaultDetector] instance from input values.
func GetFaultDetector(ctx context.Context, logger log.Logger, l1RpcApi *chain.ChainAPIClient, l2RpcApi *chain.ChainAPIClient, oracleContractAccessor OracleAccessor, faultProofWindow uint64, currentOutputIndex uint64, metrics *faultDetectorMetrics, notification *notification.Notification, diverged bool, wg *sync.WaitGroup, errorChan chan error, mutex *sync.RWMutex) *FaultDetector {
	return &FaultDetector{
		ctx:                    ctx,
		logger:                 logger,
		errorChan:              errorChan,
		wg:                     wg,
		l1RpcApi:               l1RpcApi,
		l2RpcApi:               l2RpcApi,
		oracleContractAccessor: oracleContractAccessor,
		faultProofWindow:       faultProofWindow,
		currentOutputIndex:     currentOutputIndex,
		diverged:               diverged,
		metrics:                metrics,
		notification:           notification,
		mutex:                  mutex,
	}
}
