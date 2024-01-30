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
	currentOutputIndex     int64
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
	logger.Infof("Fault proof window is set to: %d", faultProofWindow)

	var currentOutputIndex int64
	if faultDetectorConfig.Startbatchindex == -1 {
		// TODO: Get currentOutputIndexn based on FindFirstUnfinalizedStateBatchIndex output
		// logger.Infof("Finding appropriate starting unfinalized batch")
		// firstUnfinalized := FindFirstUnfinalizedStateBatchIndex(
		// 	oracleContractAccessor,
		// 	faultProofWindow,
		// 	logger,
		// )

		// if !firstUnfinalized {
		// 	logger.Infof("no unfinalized batches found. skipping all batches.")
		// 	nextOutputIndex, err := oracleContractAccessor.GetNextOutputIndex()
		// 	if err != nil {
		// 		logger.Errorf("Failed to query next output index %s", err)
		// 		return nil, err
		// 	}
		// 	currentOutputIndex = encoding.MustConvertBigIntToUint64(nextOutputIndex) - 1
		// } else {
		// 	currentBatchIndex = firstUnfinalized
		// }
	} else {
		currentOutputIndex = faultDetectorConfig.Startbatchindex
	}

	logger.Infof("Starting batch is set to: %d", currentOutputIndex)

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
	fd.logger.Infof("Checking current batch with output index: %d", fd.currentOutputIndex)

	nextOutputIndex, err := fd.oracleContractAccessor.GetNextOutputIndex()
	if err != nil {
		fd.logger.Errorf("Failed to query next output index")
		return
	}

	currentOutputIndex := uint64(fd.currentOutputIndex)
	latestBatchIndex := encoding.MustConvertBigIntToUint64(nextOutputIndex) - 1
	if currentOutputIndex > latestBatchIndex {
		fd.logger.Infof("Current output index %d is ahead of the oracle latest batch index %d", fd.currentOutputIndex, latestBatchIndex)
		return
	}

	l2OutputData, err := fd.oracleContractAccessor.GetL2Output(encoding.MustConvertUint64ToBigInt(currentOutputIndex))
	if err != nil {
		fd.logger.Errorf("Failed to fetch output associated with index %d", fd.currentOutputIndex)
		return
	}

	latestBlockNumber, err := fd.l2RpcApi.GetLatestBlockNumber(fd.ctx)
	if err != nil {
		fd.logger.Errorf("Failed to query L2 latest block number %s", err)
		return
	}

	l2OutputBlockNumber := l2OutputData.L2BlockNumber
	if latestBlockNumber < l2OutputBlockNumber {
		fd.logger.Errorf("L2 node is behind, waiting for node to sync with the network...")
		return
	}

	outputBlock, err := fd.l2RpcApi.GetBlockByNumber(fd.ctx, encoding.MustConvertUint64ToBigInt(l2OutputBlockNumber))
	if err != nil {
		fd.logger.Errorf("Failed to fetch output block by number: %d", l2OutputBlockNumber)
		return
	}

	messagePasserProofResponse, err := fd.l2RpcApi.GetProof(fd.ctx, encoding.MustConvertUint64ToBigInt(l2OutputBlockNumber), common.HexToAddress(chain.DefaultL2ContractAddresses.BedrockMessagePasser))
	if err != nil {
		fd.logger.Errorf("Failed to fetch message passer proof for the block %d and address %s", l2OutputBlockNumber, chain.DefaultL2ContractAddresses.BedrockMessagePasser)
		return
	}

	outputRoot := encoding.ComputeL2OutputRoot(
		outputBlock.Root(),
		messagePasserProofResponse.StorageHash,
		outputBlock.Hash(),
	)

	if outputRoot != common.HexToHash(l2OutputData.OutputRoot) {
		fd.diverged = true
		fd.logger.Errorf("State root does not match expectedStateRoot: %s, calculatedStateRoot: %s", l2OutputData.OutputRoot, outputRoot)
		return
	}

	fd.logger.Infof("Successfully checked current batch with index %d --> ok", fd.currentOutputIndex)
	fd.diverged = false
	fd.currentOutputIndex++
}
