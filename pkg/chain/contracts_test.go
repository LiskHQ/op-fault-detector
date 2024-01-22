package chain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetL1OracleContractAddressByChainID(t *testing.T) {
	assert := assert.New(t)

	const availableChainID uint64 = 10
	oracleContractAddressExpected := "0xdfe97868233d1aa22e815a266982f2cf17685a27"
	contractAddress, err := getL1OracleContractAddressByChainID(availableChainID)
	assert.NoError(err)
	assert.Equal(oracleContractAddressExpected, contractAddress)

	const unavailableChainID uint64 = 5
	contractAddress, err = getL1OracleContractAddressByChainID(unavailableChainID)
	assert.Error(err)
	assert.Equal(0, len(contractAddress))
}
