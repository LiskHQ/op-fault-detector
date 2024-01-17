package chain

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/LiskHQ/op-fault-detector/pkg/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
)

const (
	fake_url = "localhost:8080"
)

func TestNewChainAPIClient(t *testing.T) {
	log := log.DefaultLogger

	apiClientMock := &MockAPIClient{}
	chainClient, _ := NewChainAPIClient(apiClientMock, log)
	assert.Equal(t, chainClient.log, log)
}

func TestGetChainID(t *testing.T) {
	ctx := context.Background()
	log := log.DefaultLogger
	apiClientMock := &MockAPIClient{}
	chainClient, _ := NewChainAPIClient(apiClientMock, log)

	assert.Equal(t, chainClient.log, log)

	expectedChainID := big.NewInt(2)
	apiClientMock.On("ChainID", context.Background()).Return(expectedChainID, nil)
	receivedChainID, _ := chainClient.GetChainID(ctx)
	assert.Equal(t, expectedChainID, receivedChainID)
}

func TestGetLatestBlockNumber(t *testing.T) {
	ctx := context.Background()
	log := log.DefaultLogger
	apiClientMock := &MockAPIClient{}
	chainClient, _ := NewChainAPIClient(apiClientMock, log)

	assert.Equal(t, chainClient.log, log)

	expectedLatestBlockNumber := uint64(12345)
	apiClientMock.On("BlockNumber", context.Background()).Return(expectedLatestBlockNumber, nil)
	receivedLatestBlockNumber, _ := chainClient.GetLatestBlockNumber(ctx)
	assert.Equal(t, expectedLatestBlockNumber, receivedLatestBlockNumber)
}

func TestGetBlockByNumber(t *testing.T) {
	ctx := context.Background()
	log := log.DefaultLogger
	apiClientMock := &MockAPIClient{}
	chainClient, _ := NewChainAPIClient(apiClientMock, log)

	assert.Equal(t, chainClient.log, log)

	blockNumber := big.NewInt(800)
	expectedBlock := &types.Block{
		ReceivedAt: time.Now(),
	}
	apiClientMock.On("BlockByNumber", context.Background(), blockNumber).Return(expectedBlock, nil)
	receivedBlock, _ := chainClient.GetBlockByNumber(ctx, blockNumber)
	assert.Equal(t, expectedBlock, receivedBlock)
}

func TestGetProof(t *testing.T) {
	// t.Skipf("Skipping GetProof test")
	log := log.DefaultLogger
	apiClientMock := &MockAPIClient{}
	chainClient, _ := NewChainAPIClient(apiClientMock, log)

	assert.Equal(t, chainClient.log, log)

	rpcClientMock := &MockRPCClient{}
	// Args for Call() method
	address := common.Address{}
	blockNumber := big.NewInt(800)
	var proofResponseExpected ProofResponse

	apiClientMock.On("Client").Return(rpcClientMock, nil)
	rpcClientMock.On("Call", &proofResponseExpected, RPCEndpointGetProof, address.String(), []string{}, blockNumber).Return(fmt.Println("Error"))

	proofResponseRecieved, _ := chainClient.GetProof(rpcClientMock, blockNumber, address.Hex())

	assert.Equal(t, proofResponseExpected, proofResponseRecieved)
}
