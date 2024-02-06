package faultdetector

import (
	"context"
	crand "crypto/rand"
	"fmt"
	"math/big"
	"testing"

	"github.com/LiskHQ/op-fault-detector/pkg/chain"
	"github.com/LiskHQ/op-fault-detector/pkg/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockChainAPIClient struct {
	mock.Mock
}
type mockOracleAccessor struct {
	mock.Mock
}

func (m *mockChainAPIClient) GetLatestBlockHeader(ctx context.Context) (*types.Header, error) {
	called := m.MethodCalled("GetLatestBlockHeader", ctx)
	return called.Get(0).(*types.Header), called.Error(1)
}

func (o *mockOracleAccessor) GetNextOutputIndex() (*big.Int, error) {
	called := o.MethodCalled("GetNextOutputIndex")
	return called.Get(0).(*big.Int), called.Error(1)
}

func (o *mockOracleAccessor) GetL2Output(index *big.Int) (chain.L2Output, error) {
	called := o.MethodCalled("GetL2Output", index)
	return called.Get(0).(chain.L2Output), called.Error(1)
}

func randHash() (out common.Hash) {
	_, _ = crand.Read(out[:])
	return out
}

func TestFindFirstUnfinalizedOutputIndex(t *testing.T) {
	const defaultL1Timestamp uint64 = 123456
	const finalizationPeriodSeconds uint64 = 1000
	var hdr = &types.Header{
		ParentHash:  randHash(),
		UncleHash:   randHash(),
		Coinbase:    common.Address{},
		Root:        randHash(),
		TxHash:      randHash(),
		ReceiptHash: randHash(),
		Bloom:       types.Bloom{},
		Difficulty:  big.NewInt(42),
		Number:      big.NewInt(1234),
		GasLimit:    0,
		GasUsed:     0,
		Time:        defaultL1Timestamp + finalizationPeriodSeconds,
		Extra:       make([]byte, 0),
		MixDigest:   randHash(),
		Nonce:       types.BlockNonce{},
		BaseFee:     big.NewInt(100),
	}
	type testSuite struct {
		l2Client *mockChainAPIClient
		oracle   *mockOracleAccessor
		logger   log.Logger
		ctx      context.Context
	}

	var tests = []struct {
		name         string
		construction func() *testSuite
		assertion    func(uint64, error)
	}{
		{
			name: "when fails to get latest block header from L2 provider",
			construction: func() *testSuite {
				l2Client := new(mockChainAPIClient)
				oracle := new(mockOracleAccessor)
				logger, _ := log.NewDefaultProductionLogger()
				ctx := context.Background()

				sampleL2Output1 := chain.L2Output{
					OutputRoot:    randHash().String(),
					L1Timestamp:   defaultL1Timestamp - 1,
					L2BlockNumber: 500,
					L2OutputIndex: 2,
				}
				sampleL2Output2 := chain.L2Output{
					OutputRoot:    randHash().String(),
					L1Timestamp:   defaultL1Timestamp + 1,
					L2BlockNumber: 500,
					L2OutputIndex: 2,
				}

				l2Client.On("GetLatestBlockHeader", ctx).Return(hdr, fmt.Errorf("Failed to get latest block header"))
				oracle.On("GetNextOutputIndex").Return(big.NewInt(2), nil)
				oracle.On("GetL2Output", big.NewInt(0)).Return(sampleL2Output1, nil)
				oracle.On("GetL2Output", big.NewInt(1)).Return(sampleL2Output2, nil)

				return &testSuite{
					l2Client: l2Client,
					oracle:   oracle,
					logger:   logger,
					ctx:      ctx,
				}
			},
			assertion: func(index uint64, err error) {
				require.EqualError(t, err, "Failed to get latest block header")
				var expected uint64 = 0
				require.Equal(t, index, expected)
			},
		},
		{
			name: "when fails to get next output index",
			construction: func() *testSuite {
				l2Client := new(mockChainAPIClient)
				oracle := new(mockOracleAccessor)
				logger, _ := log.NewDefaultProductionLogger()
				ctx := context.Background()

				sampleL2Output1 := chain.L2Output{
					OutputRoot:    randHash().String(),
					L1Timestamp:   defaultL1Timestamp - 1,
					L2BlockNumber: 500,
					L2OutputIndex: 2,
				}
				sampleL2Output2 := chain.L2Output{
					OutputRoot:    randHash().String(),
					L1Timestamp:   defaultL1Timestamp + 1,
					L2BlockNumber: 500,
					L2OutputIndex: 2,
				}

				l2Client.On("GetLatestBlockHeader", ctx).Return(hdr, nil)
				oracle.On("GetNextOutputIndex").Return(big.NewInt(2), fmt.Errorf("Failed to get next output index"))
				oracle.On("GetL2Output", big.NewInt(0)).Return(sampleL2Output1, nil)
				oracle.On("GetL2Output", big.NewInt(1)).Return(sampleL2Output2, nil)

				return &testSuite{
					l2Client: l2Client,
					oracle:   oracle,
					logger:   logger,
					ctx:      ctx,
				}
			},
			assertion: func(index uint64, err error) {
				require.EqualError(t, err, "Failed to get next output index")
				var expected uint64 = 0
				require.Equal(t, index, expected)
			},
		},
		{
			name: "when fails to get L2 output",
			construction: func() *testSuite {
				l2Client := new(mockChainAPIClient)
				oracle := new(mockOracleAccessor)
				logger, _ := log.NewDefaultProductionLogger()
				ctx := context.Background()

				sampleL2Output1 := chain.L2Output{
					OutputRoot:    randHash().String(),
					L1Timestamp:   defaultL1Timestamp - 1,
					L2BlockNumber: 500,
					L2OutputIndex: 2,
				}
				sampleL2Output2 := chain.L2Output{
					OutputRoot:    randHash().String(),
					L1Timestamp:   defaultL1Timestamp + 1,
					L2BlockNumber: 500,
					L2OutputIndex: 2,
				}

				l2Client.On("GetLatestBlockHeader", ctx).Return(hdr, nil)
				oracle.On("GetNextOutputIndex").Return(big.NewInt(2), nil)
				oracle.On("GetL2Output", big.NewInt(0)).Return(sampleL2Output1, fmt.Errorf("Failed to get L2 output"))
				oracle.On("GetL2Output", big.NewInt(1)).Return(sampleL2Output2, nil)

				return &testSuite{
					l2Client: l2Client,
					oracle:   oracle,
					logger:   logger,
					ctx:      ctx,
				}
			},
			assertion: func(index uint64, err error) {
				require.EqualError(t, err, "Failed to get L2 output")
				var expected uint64 = 0
				require.Equal(t, index, expected)
			},
		},
		{
			name: "when the chain is more then FPW seconds old",
			construction: func() *testSuite {
				l2Client := new(mockChainAPIClient)
				oracle := new(mockOracleAccessor)
				logger, _ := log.NewDefaultProductionLogger()
				ctx := context.Background()

				sampleL2Output1 := chain.L2Output{
					OutputRoot:    randHash().String(),
					L1Timestamp:   defaultL1Timestamp - 1,
					L2BlockNumber: 500,
					L2OutputIndex: 2,
				}
				sampleL2Output2 := chain.L2Output{
					OutputRoot:    randHash().String(),
					L1Timestamp:   defaultL1Timestamp + 1,
					L2BlockNumber: 500,
					L2OutputIndex: 2,
				}

				l2Client.On("GetLatestBlockHeader", ctx).Return(hdr, nil)
				oracle.On("GetNextOutputIndex").Return(big.NewInt(2), nil)
				oracle.On("GetL2Output", big.NewInt(0)).Return(sampleL2Output1, nil)
				oracle.On("GetL2Output", big.NewInt(1)).Return(sampleL2Output2, nil)

				return &testSuite{
					l2Client: l2Client,
					oracle:   oracle,
					logger:   logger,
					ctx:      ctx,
				}
			},
			assertion: func(index uint64, err error) {
				require.NoError(t, err)
				var expected uint64 = 1
				require.Equal(t, index, expected)
			},
		},
		{
			name: "when the chain is less than FPW seconds old",
			construction: func() *testSuite {
				l2Client := new(mockChainAPIClient)
				oracle := new(mockOracleAccessor)
				logger, _ := log.NewDefaultProductionLogger()
				ctx := context.Background()

				sampleL2Output1 := chain.L2Output{
					OutputRoot:    randHash().String(),
					L1Timestamp:   defaultL1Timestamp + 1,
					L2BlockNumber: 500,
					L2OutputIndex: 2,
				}
				sampleL2Output2 := chain.L2Output{
					OutputRoot:    randHash().String(),
					L1Timestamp:   defaultL1Timestamp + 2,
					L2BlockNumber: 500,
					L2OutputIndex: 2,
				}

				l2Client.On("GetLatestBlockHeader", ctx).Return(hdr, nil)
				oracle.On("GetNextOutputIndex").Return(big.NewInt(2), nil)
				oracle.On("GetL2Output", big.NewInt(0)).Return(sampleL2Output1, nil)
				oracle.On("GetL2Output", big.NewInt(1)).Return(sampleL2Output2, nil)

				return &testSuite{
					l2Client: l2Client,
					oracle:   oracle,
					logger:   logger,
					ctx:      ctx,
				}
			},
			assertion: func(index uint64, err error) {
				require.NoError(t, err)
				var expected uint64 = 0
				require.Equal(t, index, expected)
			},
		},
		{
			name: "when no batches submitted for the entire FPW",
			construction: func() *testSuite {
				l2Client := new(mockChainAPIClient)
				oracle := new(mockOracleAccessor)
				logger, _ := log.NewDefaultProductionLogger()
				ctx := context.Background()

				sampleL2Output1 := chain.L2Output{
					OutputRoot:    randHash().String(),
					L1Timestamp:   defaultL1Timestamp - 2,
					L2BlockNumber: 500,
					L2OutputIndex: 2,
				}
				sampleL2Output2 := chain.L2Output{
					OutputRoot:    randHash().String(),
					L1Timestamp:   defaultL1Timestamp - 1,
					L2BlockNumber: 500,
					L2OutputIndex: 2,
				}

				l2Client.On("GetLatestBlockHeader", ctx).Return(hdr, nil)
				oracle.On("GetNextOutputIndex").Return(big.NewInt(2), nil)
				oracle.On("GetL2Output", big.NewInt(0)).Return(sampleL2Output1, nil)
				oracle.On("GetL2Output", big.NewInt(1)).Return(sampleL2Output2, nil)

				return &testSuite{
					l2Client: l2Client,
					oracle:   oracle,
					logger:   logger,
					ctx:      ctx,
				}
			},
			assertion: func(index uint64, err error) {
				require.EqualError(t, err, "Undefined")
				var expected uint64 = 0
				require.Equal(t, index, expected)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := test.construction()

			index, err := FindFirstUnfinalizedOutputIndex(ts.ctx, ts.logger, finalizationPeriodSeconds, ts.oracle, ts.l2Client)
			test.assertion(index, err)
		})
	}
}
