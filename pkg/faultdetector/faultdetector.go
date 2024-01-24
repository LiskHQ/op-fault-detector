// Package faultdetector implements Optimism fault detector.
package faultdetector

import (
	"context"
	"sync"
	"time"

	"github.com/LiskHQ/op-fault-detector/pkg/chain"
	"github.com/LiskHQ/op-fault-detector/pkg/config"
	"github.com/LiskHQ/op-fault-detector/pkg/log"
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

	// TODO: Calculate from findFirstUnfinalizedOutputIndex(context, OracleContractAccessor, L1Provider, faultProofWindow, logger)

	faultDetector := &FaultDetector{
		ctx:                    ctx,
		logger:                 logger,
		errorChan:              errorChan,
		wg:                     wg,
		l1RpcApi:               l1RpcApi,
		l2RpcApi:               l2RpcApi,
		oracleContractAccessor: oracleContractAccessor,
		faultProofWindow:       uint64(2), // TODO
		currentOutputIndex:     uint64(2), // TODO
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

// TODO: Implement checkFault to check for faults
func (fd *FaultDetector) checkFault() {
	// TODO: The below log need to be removed/updated after full implementation
	fd.logger.Infof("Connected to L1 and L2 chains, the currentOutputIndex is %d", fd.currentOutputIndex)
}
