// Package chain implements everything related to interaction with smart contracts, rpcprovider, etc
package chain

import (
	"context"
	"math/big"

	"github.com/LiskHQ/op-fault-detector/pkg/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

type ChainAPIClient struct {
	client *APIClient
	log    log.Logger
}

// NewChainAPIClient returns a [ChainAPIClient], wrapping all RPC endpoints to access chain related data.
func NewChainAPIClient(log log.Logger) (*ChainAPIClient, error) {
	return &ChainAPIClient{
		log: log,
	}, nil
}

func (c *ChainAPIClient) Connect(ethClientObj EthClientInterface, url string) {
	client, err := ethClientObj.Dial(url)
	if err != nil {
		c.log.Errorf("Error occurred while connecting %w", err)
	}
	c.client = NewAPIClient(client)
}

// Returns chainID of the connected node.
func (c *ChainAPIClient) GetChainID() (*big.Int, error) {
	chainID, err := c.client.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	return chainID, nil
}

// Returns latest block number from the connected node.
func (c *ChainAPIClient) GetLatestBlockNumber() (uint64, error) {
	blockNumber, err := c.client.BlockNumber(context.Background())
	if err != nil {
		return 0, err
	}

	return blockNumber, nil
}

// Returns block for a given block number from the connected node.
func (c *ChainAPIClient) GetBlockByNumber(blockNumber *big.Int) (*types.Block, error) {
	block, err := c.client.BlockByNumber(context.Background(), blockNumber)
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

const (
	RPCEndpointGetProof = "eth_getProof"
)

// Returns the account and storage values, including the Merkle proof, of the specified account/address.
func (c *ChainAPIClient) GetProof(blockNumber *big.Int, address string) (*ProofResponse, error) {
	var result ProofResponse

	rpcClient := &RPCClient{c.client}
	if err := rpcClient.Call(&result, RPCEndpointGetProof, address, []string{}, hexutil.EncodeBig(blockNumber)); err != nil {
		return nil, err
	}

	return &result, nil
}
