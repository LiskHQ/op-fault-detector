// Package faultdetector implements Optimism fault detector.
package faultdetector

import (
	"context"
	"sync"
	"time"

	"github.com/LiskHQ/op-fault-detector/pkg/chain"
	"github.com/LiskHQ/op-fault-detector/pkg/config"
	"github.com/LiskHQ/op-fault-detector/pkg/encoding"
	"github.com/LiskHQ/op-fault-detector/pkg/log"
	"github.com/ethereum/go-ethereum/common"
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
	l1RpcApi               *chain.ChainAPIClient
	l2RpcApi               *chain.ChainAPIClient
	oracleContractAccessor *chain.OracleAccessor
	faultProofWindow       uint64
	currentOutputIndex     uint64
	diverged               bool
	ticker                 *time.Ticker
	quitTickerChan         chan struct{}
}

// NewFaultDetector will return [FaultDetector] with the initialized providers and configuration.
func NewFaultDetector(ctx context.Context, logger log.Logger, errorChan chan error, wg *sync.WaitGroup, faultDetectorConfig *config.FaultDetectorConfig) (*FaultDetector, error) {
	// Initialize API Providers
	l1RpcApi, err := chain.GetAPIClient(ctx, faultDetectorConfig.L1RPCEndpoint, logger)
	if err != nil {
		logger.Errorf("Failed to create API client for L1 Provider with given endpoint: %s, error: %w", faultDetectorConfig.L1RPCEndpoint, err)
		return nil, err
	}

	l2RpcApi, err := chain.GetAPIClient(ctx, faultDetectorConfig.L2RPCEndpoint, logger)
	if err != nil {
		logger.Errorf("Failed to create API client for L2 Provider with given endpoint: %s, error: %w", faultDetectorConfig.L2RPCEndpoint, err)
		return nil, err
	}

	l2ChainID, err := l2RpcApi.GetChainID(ctx)
	if err != nil {
		logger.Errorf("Failed to get L2 provider's chainID: %d, error: %w", l2ChainID.Int64(), err)
		return nil, err
	}

	// Initialize Oracle contract accessor
	chainConfig := &chain.ConfigOptions{
		L1RPCEndpoint:                 faultDetectorConfig.L1RPCEndpoint,
		ChainID:                       l2ChainID.Uint64(),
		L2OutputOracleContractAddress: faultDetectorConfig.L2OutputOracleContractAddress,
	}

	oracleContractAccessor, err := chain.NewOracleAccessor(ctx, chainConfig)

	if err != nil {
		logger.Errorf("Failed to create Oracle contract accessor with chainID: %d, L1 endpoint: %s and L2OutputOracleContractAddress: %s, error: %w", l2ChainID.Int64(), faultDetectorConfig.L1RPCEndpoint, faultDetectorConfig.L2OutputOracleContractAddress, err)
		return nil, err
	}

	faultProofWindow, err := oracleContractAccessor.FinalizationPeriodSeconds()
	if err != nil {
		logger.Errorf("Failed to query FinalizationPeriodSecond", err)
		return nil, err
	}
	logger.Infof("Fault proof window is set to %d.", faultProofWindow)

	var currentOutputIndex uint64
	if int64(faultDetectorConfig.Startbatchindex) == -1 {
		logger.Infof("Finding appropriate starting unfinalized batch....")
		// firstUnfinalized, _ := FindFirstUnfinalizedOutputIndex(
		// 	ctx,
		// 	logger,
		// 	encoding.MustConvertBigIntToUint64(faultProofWindow),
		// 	oracleContractAccessor,
		// 	l2RpcApi,
		// )
		// if firstUnfinalized == 0 {
		// 	logger.Infof("No unfinalized batches found. skipping all batches.")
		// 	nextOutputIndex, err := oracleContractAccessor.GetNextOutputIndex()
		// 	if err != nil {
		// 		logger.Errorf("Failed to query next output index %s", err)
		// 		return nil, err
		// 	}
		// 	currentOutputIndex = encoding.MustConvertBigIntToUint64(nextOutputIndex) - 1
		// } else {
		// 	currentOutputIndex = firstUnfinalized
		// }
	} else {
		currentOutputIndex = faultDetectorConfig.Startbatchindex
	}
	logger.Infof("Starting unfinalized batch index is set to %d.", currentOutputIndex)

	faultDetector := &FaultDetector{
		ctx:                    ctx,
		logger:                 logger,
		errorChan:              errorChan,
		wg:                     wg,
		l1RpcApi:               l1RpcApi,
		l2RpcApi:               l2RpcApi,
		oracleContractAccessor: oracleContractAccessor,
		faultProofWindow:       encoding.MustConvertBigIntToUint64(faultProofWindow),
		currentOutputIndex:     currentOutputIndex,
		diverged:               false,
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
			fd.checkFault()
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
func (fd *FaultDetector) checkFault() {
	startTime := time.Now()

	fd.logger.Infof("Checking current batch with output index %d.", fd.currentOutputIndex)

	nextOutputIndex, err := fd.oracleContractAccessor.GetNextOutputIndex()
	if err != nil {
		fd.logger.Errorf("Failed to query next output index.")
		time.Sleep(waitTimeInFailure * time.Second)
		return
	}

	latestBatchIndex := encoding.MustConvertBigIntToUint64(nextOutputIndex) - 1
	if fd.currentOutputIndex > latestBatchIndex {
		fd.logger.Infof("Current output index %d is ahead of the oracle latest batch index %d. Waiting...", fd.currentOutputIndex, latestBatchIndex)
		time.Sleep(waitTimeInFailure * time.Second)
		return
	}

	l2OutputData, err := fd.oracleContractAccessor.GetL2Output(encoding.MustConvertUint64ToBigInt(fd.currentOutputIndex))
	if err != nil {
		fd.logger.Errorf("Failed to fetch output associated with index %d.", fd.currentOutputIndex)
		time.Sleep(waitTimeInFailure * time.Second)
		return
	}

	latestBlockNumber, err := fd.l2RpcApi.GetLatestBlockNumber(fd.ctx)
	if err != nil {
		fd.logger.Errorf("Failed to query L2 latest block number %d", latestBlockNumber)
		time.Sleep(waitTimeInFailure * time.Second)
		return
	}

	l2OutputBlockNumber := l2OutputData.L2BlockNumber
	expectedOutputRoot := l2OutputData.OutputRoot
	if latestBlockNumber < l2OutputBlockNumber {
		fd.logger.Errorf("L2 node is behind, waiting for node to sync with the network...")
		time.Sleep(waitTimeInFailure * time.Second)
		return
	}

	outputBlockHeader, err := fd.l2RpcApi.GetBlockHeaderByNumber(fd.ctx, encoding.MustConvertUint64ToBigInt(l2OutputBlockNumber))
	if err != nil {
		fd.logger.Errorf("Failed to fetch block header by number %d.", l2OutputBlockNumber)
		time.Sleep(waitTimeInFailure * time.Second)
		return
	}

	messagePasserProofResponse, err := fd.l2RpcApi.GetProof(fd.ctx, encoding.MustConvertUint64ToBigInt(l2OutputBlockNumber), common.HexToAddress(chain.DefaultL2ContractAddresses.BedrockMessagePasser))
	if err != nil {
		fd.logger.Errorf("Failed to fetch message passer proof for the block %d and address %s.", l2OutputBlockNumber, chain.DefaultL2ContractAddresses.BedrockMessagePasser)
		time.Sleep(waitTimeInFailure * time.Second)
		return
	}

	calculatedOutputRoot := encoding.ComputeL2OutputRoot(
		outputBlockHeader.Root,
		messagePasserProofResponse.StorageHash,
		outputBlockHeader.Hash(),
	)

	if calculatedOutputRoot != expectedOutputRoot {
		fd.diverged = true
		finalizationTime := time.Unix(int64(outputBlockHeader.Time+fd.faultProofWindow), 0)
		fd.logger.Errorf("State root does not match expectedStateRoot: %s, calculatedStateRoot: %s, finalizationTime: %s.", expectedOutputRoot, calculatedOutputRoot, finalizationTime)
		return
	}

	// Time taken to execute each batch in milliseconds.
	elapsedTime := time.Since(startTime).Milliseconds()
	fd.logger.Infof("Successfully checked current batch with index %d --> ok, time taken %dms.", fd.currentOutputIndex, elapsedTime)
	fd.diverged = false
	fd.currentOutputIndex++
}
