// Package chain implements everything related to interaction with smart contracts, rpcprovider, etc
package chain

import (
	"context"
	"math/big"

	"github.com/LiskHQ/op-fault-detector/pkg/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

const (
	RPCEndpointGetProof = "eth_getProof"
)

type APIMethods interface {
	ChainID(ctx context.Context) (*big.Int, error)
	BlockNumber(ctx context.Context) (uint64, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
	Client() *rpc.Client
}

type RPCClient interface {
	Call(result interface{}, method string, args ...interface{}) error
}

// GetAPIClient return [ChainAPIClient] with client attached.
func GetAPIClient(url string, log log.Logger) (*ChainAPIClient, error) {
	client, err := ethclient.Dial(url)
	if err != nil {
		return nil, err
	}

	return NewChainAPIClient(client, log)
}

type ChainAPIClient struct {
	apiClient APIMethods
	log       log.Logger
}

// NewChainAPIClient returns a [ChainAPIClient], wrapping all RPC endpoints to access chain related data.
func NewChainAPIClient(apiClient APIMethods, log log.Logger) (*ChainAPIClient, error) {
	return &ChainAPIClient{
		apiClient: apiClient,
		log:       log,
	}, nil
}

// Returns chainID of the connected node.
func (c *ChainAPIClient) GetChainID(ctx context.Context) (*big.Int, error) {
	chainID, err := c.apiClient.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	return chainID, nil
}

// Returns latest block number from the connected node.
func (c *ChainAPIClient) GetLatestBlockNumber(ctx context.Context) (uint64, error) {
	blockNumber, err := c.apiClient.BlockNumber(ctx)
	if err != nil {
		return 0, err
	}

	return blockNumber, nil
}

// Returns block for a given block number from the connected node.
func (c *ChainAPIClient) GetBlockByNumber(ctx context.Context, blockNumber *big.Int) (*types.Block, error) {
	block, err := c.apiClient.BlockByNumber(ctx, blockNumber)
	if err != nil {
		return nil, err
	}

	return block, nil
}

type ProofResponse struct {
	Address      common.Address  `json:"address"`
	AccountProof []hexutil.Bytes `json:"accountProof"`
	Balance      *hexutil.Big    `json:"balance"`
	CodeHash     common.Hash     `json:"codeHash"`
	Nonce        hexutil.Uint64  `json:"nonce"`
	StorageHash  common.Hash     `json:"storageHash"`
	StorageProof []common.Hash   `json:"storageProof"`
}

// Returns the account and storage values, including the Merkle proof, of the specified account/address.
func (c *ChainAPIClient) GetProof(client RPCClient, blockNumber *big.Int, address string) (*ProofResponse, error) {
	var result ProofResponse

	if err := client.Call(&result, RPCEndpointGetProof, address, []string{}, blockNumber); err != nil {
		return nil, err
	}

	return &result, nil
}
