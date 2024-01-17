package chain

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type EthClientInterface interface {
	Dial(url string) (APIClientInterface, error)
}

type EthClient struct{}

func (e *EthClient) Dial(url string) (APIClientInterface, error) {
	client, err := ethclient.Dial(url)
	if err != nil {
		return nil, err
	}

	return client, nil
}

type APIClientInterface interface {
	ChainID(ctx context.Context) (*big.Int, error)
	BlockNumber(ctx context.Context) (uint64, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
	Client() *rpc.Client
}

type APIClient struct {
	client APIClientInterface
}

func NewAPIClient(client APIClientInterface) *APIClient {
	return &APIClient{client: client}
}
func (c *APIClient) ChainID(ctx context.Context) (*big.Int, error) {
	return c.client.ChainID(context.Background())
}
func (c *APIClient) BlockNumber(ctx context.Context) (uint64, error) {
	return c.client.BlockNumber(context.Background())
}
func (c *APIClient) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	return c.client.BlockByNumber(context.Background(), number)
}
func (c *APIClient) Client() *rpc.Client {
	return c.client.Client()
}

type RPCClientInterface interface {
	Call(result interface{}, method string, args ...interface{}) error
}

type RPCClient struct {
	apiClient APIClientInterface
}

func (c *RPCClient) Call(result interface{}, method string, args ...interface{}) error {
	return c.apiClient.Client().Call(result, method, args)
}
