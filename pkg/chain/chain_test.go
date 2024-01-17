package chain

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/LiskHQ/op-fault-detector/pkg/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
)

const (
	fake_url = "localhost:8080"
)

func TestNewChainAPIClient(t *testing.T) {
	log := log.DefaultLogger

	chainClient, _ := NewChainAPIClient(log)
	assert.Equal(t, chainClient.log, log)
}

func TestConnect(t *testing.T) {
	log := log.DefaultLogger
	ethClientMock := &MockEthClient{}

	apiClientMock := &MockAPIClient{}
	ethClientMock.On("Dial", fake_url).Return(apiClientMock, nil)

	chainClient, _ := NewChainAPIClient(log)
	chainClient.Connect(ethClientMock, fake_url)

	assert.Equal(t, chainClient.log, log)
}

func TestGetChainID(t *testing.T) {
	log := log.DefaultLogger
	ethClientMock := &MockEthClient{}

	apiClientMock := &MockAPIClient{}
	ethClientMock.On("Dial", fake_url).Return(apiClientMock, nil)

	chainClient, _ := NewChainAPIClient(log)
	chainClient.Connect(ethClientMock, fake_url)

	assert.Equal(t, chainClient.log, log)

	expectedChainID := big.NewInt(2)
	apiClientMock.On("ChainID", context.Background()).Return(expectedChainID, nil)
	receivedChainID, _ := chainClient.GetChainID()
	assert.Equal(t, expectedChainID, receivedChainID)
}

func TestGetLatestBlockNumber(t *testing.T) {
	log := log.DefaultLogger
	ethClientMock := &MockEthClient{}

	apiClientMock := &MockAPIClient{}
	ethClientMock.On("Dial", fake_url).Return(apiClientMock, nil)

	chainClient, _ := NewChainAPIClient(log)
	chainClient.Connect(ethClientMock, fake_url)

	assert.Equal(t, chainClient.log, log)

	expectedLatestBlockNumber := uint64(12345)
	apiClientMock.On("BlockNumber", context.Background()).Return(expectedLatestBlockNumber, nil)
	receivedLatestBlockNumber, _ := chainClient.GetLatestBlockNumber()
	assert.Equal(t, expectedLatestBlockNumber, receivedLatestBlockNumber)
}

func TestGetBlockByNumber(t *testing.T) {
	log := log.DefaultLogger
	ethClientMock := &MockEthClient{}

	apiClientMock := &MockAPIClient{}
	ethClientMock.On("Dial", fake_url).Return(apiClientMock, nil)

	chainClient, _ := NewChainAPIClient(log)
	chainClient.Connect(ethClientMock, fake_url)

	assert.Equal(t, chainClient.log, log)

	blockNumber := big.NewInt(800)
	expectedBlock := &types.Block{
		ReceivedAt: time.Now(),
	}
	apiClientMock.On("BlockByNumber", context.Background(), blockNumber).Return(expectedBlock, nil)
	receivedBlock, _ := chainClient.GetBlockByNumber(blockNumber)
	assert.Equal(t, expectedBlock, receivedBlock)
}

func TestGetProof(t *testing.T) {
	t.Skip("skipping testing")
	log := log.DefaultLogger
	ethClientMock := &MockEthClient{}

	apiClientMock := &MockAPIClient{}
	ethClientMock.On("Dial", fake_url).Return(apiClientMock, nil)

	chainClient, _ := NewChainAPIClient(log)
	chainClient.Connect(ethClientMock, fake_url)

	assert.Equal(t, chainClient.log, log)

	rpcClientMock := &MockRPCClient{}
	apiClientMock.On("Client").Return(rpcClientMock, nil)

	address := common.Address{}
	blockNumber := big.NewInt(800)
	proofResponseExpected := &ProofResponse{}
	rpcClientMock.On("Call", proofResponseExpected, RPCEndpointGetProof, address, []string{}, hexutil.EncodeBig(blockNumber)).Return(nil)
	proofResponseRecieved, _ := chainClient.GetProof(blockNumber, address.Hex())
	assert.Equal(t, proofResponseExpected, proofResponseRecieved)
}
