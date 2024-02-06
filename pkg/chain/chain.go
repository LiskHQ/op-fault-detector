// Package chain implements everything related to interaction with smart contracts, rpc provider, etc.
package chain

import (
	"context"
	"math/big"

	"github.com/LiskHQ/op-fault-detector/pkg/encoding"
	"github.com/LiskHQ/op-fault-detector/pkg/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	rpcEndpointGetProof = "eth_getProof"
)

// ChainAPIClient connects and encapsulates all the methods to interact with a chain.
type ChainAPIClient struct {
	eth *ethclient.Client
	log log.Logger
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

// GetAPIClient returns [ChainAPIClient] with client attached.
func GetAPIClient(ctx context.Context, url string, log log.Logger) (*ChainAPIClient, error) {
	client, err := ethclient.DialContext(ctx, url)
	if err != nil {
		return nil, err
	}

	return &ChainAPIClient{
		eth: client,
		log: log,
	}, nil
}

// GetChainID returns chainID of a connected node.
func (c *ChainAPIClient) GetChainID(ctx context.Context) (*big.Int, error) {
	return c.eth.ChainID(ctx)
}

// GetLatestBlockNumber returns latest block number from a connected node.
func (c *ChainAPIClient) GetLatestBlockNumber(ctx context.Context) (uint64, error) {
	return c.eth.BlockNumber(ctx)
}

// GetBlockByNumber returns block for a given block number from a connected node.
func (c *ChainAPIClient) GetBlockByNumber(ctx context.Context, blockNumber *big.Int) (*types.Block, error) {
	return c.eth.BlockByNumber(ctx, blockNumber)
}

// GetBlockHeaderByNumber returns block header for a given block number from a connected node.
func (c *ChainAPIClient) GetBlockHeaderByNumber(ctx context.Context, blockNumber *big.Int) (*types.Header, error) {
	return c.eth.HeaderByNumber(ctx, blockNumber)
}

// GetLatestBlockHeader returns latest block header from a connected node.
func (c *ChainAPIClient) GetLatestBlockHeader(ctx context.Context) (*types.Header, error) {
	blockNumber, err := c.eth.BlockNumber(ctx)
	if err != nil {
		return nil, err
	}
	return c.eth.HeaderByNumber(ctx, encoding.MustConvertUint64ToBigInt(blockNumber))
}

// GetProof returns the account and storage values, including the Merkle proof, of the specified account/address.
func (c *ChainAPIClient) GetProof(ctx context.Context, blockNumber *big.Int, address common.Address) (*ProofResponse, error) {
	var result ProofResponse

	if err := c.eth.Client().CallContext(ctx, &result, rpcEndpointGetProof, address, []string{}, hexutil.Big(*blockNumber)); err != nil {
		return nil, err
	}

	return &result, nil
}
