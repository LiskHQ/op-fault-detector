package faultdetector

import (
	"context"
	"fmt"
	"math/big"

	"github.com/LiskHQ/op-fault-detector/pkg/chain"
	"github.com/LiskHQ/op-fault-detector/pkg/encoding"
	"github.com/LiskHQ/op-fault-detector/pkg/log"
	"github.com/ethereum/go-ethereum/core/types"
)

type ChainAPIClient interface {
	GetLatestBlockHeader(ctx context.Context) (*types.Header, error)
}

type OracleAccessor interface {
	GetNextOutputIndex() (*big.Int, error)
	GetL2Output(index *big.Int) (chain.L2Output, error)
	FinalizationPeriodSeconds() (*big.Int, error)
}

// FindFirstUnfinalizedOutputIndex finds and returns the first L2 output index that has not yet passed the fault proof window.
func FindFirstUnfinalizedOutputIndex(ctx context.Context, logger log.Logger, fpw uint64, oracleAccessor OracleAccessor, l2RpcApi ChainAPIClient) (uint64, error) {
	latestBlockHeader, err := l2RpcApi.GetLatestBlockHeader(ctx)
	if err != nil {
		logger.Errorf("Failed to get latest block header from L2 provider, error: %w", err)
		return 0, err
	}
	totalOutputsBigInt, err := oracleAccessor.GetNextOutputIndex()
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
		outputData, err := oracleAccessor.GetL2Output(midBigInt)
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
