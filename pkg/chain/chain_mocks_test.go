package chain

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/mock"
)

type MockAPIClient struct {
	mock.Mock
}

func (c *MockAPIClient) ChainID(ctx context.Context) (*big.Int, error) {
	ret := c.Called(ctx)

	return ret.Get(0).(*big.Int), ret.Error(1)
}
func (c *MockAPIClient) BlockNumber(ctx context.Context) (uint64, error) {
	ret := c.Called(ctx)

	return ret.Get(0).(uint64), ret.Error(1)
}
func (c *MockAPIClient) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	ret := c.Called(ctx, number)

	return ret.Get(0).(*types.Block), ret.Error(1)
}
func (c *MockAPIClient) Client() *rpc.Client {
	ret := c.Called()

	return ret.Get(0).(*rpc.Client)
}

type MockRPCClient struct {
	mock.Mock
}

func (c *MockRPCClient) Call(result interface{}, method string, args ...interface{}) error {
	allArgs := []interface{}{result, method}
	allArgs = append(allArgs, args...)
	ret := c.Called(allArgs...)

	ptr := &result
	*ptr = ret.Get(0)
	return ret.Error(1)
}
