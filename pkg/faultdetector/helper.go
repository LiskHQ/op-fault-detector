package faultdetector

import (
	"context"
	"fmt"

	"github.com/LiskHQ/op-fault-detector/pkg/chain"
	"github.com/LiskHQ/op-fault-detector/pkg/config"
	"github.com/LiskHQ/op-fault-detector/pkg/encoding"
	"github.com/LiskHQ/op-fault-detector/pkg/log"
)

// FindFirstUnfinalizedOutputIndex finds and returns the first L2 output index that has not yet passed the fault proof window.
func FindFirstUnfinalizedOutputIndex(ctx context.Context, logger log.Logger, fpw uint64, faultDetectorConfig *config.FaultDetectorConfig) (uint64, error) {
	l2RpcApi, err := chain.GetAPIClient(ctx, faultDetectorConfig.L2RPCEndpoint, logger)
	if err != nil {
		logger.Errorf("Failed to create API client for L2 Provider with given endpoint: %s, error: %w", faultDetectorConfig.L2RPCEndpoint, err)
		return 0, err
	}
	l2ChainID, err := l2RpcApi.GetChainID(ctx)
	if err != nil {
		logger.Errorf("Failed to get L2 provider's chainID: %d, error: %w", encoding.MustConvertBigIntToUint64(l2ChainID), err)
		return 0, err
	}
	chainConfig := &chain.ConfigOptions{
		L1RPCEndpoint:                 faultDetectorConfig.L1RPCEndpoint,
		ChainID:                       encoding.MustConvertBigIntToUint64(l2ChainID),
		L2OutputOracleContractAddress: faultDetectorConfig.L2OutputOracleContractAddress,
	}
	oracleContractAccessor, err := chain.NewOracleAccessor(ctx, chainConfig)
	if err != nil {
		logger.Errorf("Failed to create Oracle contract accessor with chainID: %d, L1 endpoint: %s and L2OutputOracleContractAddress: %s, error: %w", encoding.MustConvertBigIntToUint64(l2ChainID), faultDetectorConfig.L1RPCEndpoint, faultDetectorConfig.L2OutputOracleContractAddress, err)
		return 0, err
	}
	latestBlockHeader, err := l2RpcApi.GetLatestBlockHeader(ctx)
	if err != nil {
		logger.Errorf("Failed to get latest block header from L2 provider, error: %w", err)
		return 0, err
	}
	totalOutputsBigInt, err := oracleContractAccessor.GetNextOutputIndex()
	if err != nil {
		logger.Errorf("Failed to get next output index, error: %w", err)
		return 0, err
	}
	totalOutputs := encoding.MustConvertBigIntToUint64(totalOutputsBigInt)

	// Perform a binary search to find the next batch that will pass the challenge period.
	var lo uint64 = 0
	hi := totalOutputs
	for lo != hi {
		mid := (lo + hi) / 2
		midBigInt := encoding.MustConvertUint64ToBigInt(mid)
		outputData, err := oracleContractAccessor.GetL2Output(midBigInt)
		if err != nil {
			logger.Errorf("Failed to get L2 output for index: %d, error: %w", midBigInt, err)
			return 0, err
		}
		if outputData.L1Timestamp+fpw < latestBlockHeader.Time {
			lo = mid + 1
		} else {
			hi = mid
		}
	}

	// Result will be zero if the chain is less than FPW seconds old. Only returns 0 with error Undefined in the
	// case that no batches have been submitted for an entire challenge period.
	if lo == totalOutputs {
		logger.Errorf("No batches have been submitted for the entire challenge period and therefore first unfinalized output index is undefined")
		return 0, fmt.Errorf("Undefined")
	} else {
		return lo, nil
	}
}
