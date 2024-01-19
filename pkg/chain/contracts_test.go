package chain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetL1OracleContractAddressByChainID_AvailableChainID(t *testing.T) {
	const availableChainID uint64 = 10
	var contractAddress, err = getL1OracleContractAddressByChainID(availableChainID)
	var oracleContractAddressExpected = "0xdfe97868233d1aa22e815a266982f2cf17685a27"

	assert.NoError(t, err)
	assert.Equal(t, oracleContractAddressExpected, contractAddress)
}

func TestGetL1OracleContractAddressByChainID_UnavailableChainID(t *testing.T) {
	const unavailableChainID uint64 = 5
	contractAddress, err := getL1OracleContractAddressByChainID(unavailableChainID)

	assert.Error(t, err)
	assert.Equal(t, 0, len(contractAddress))
}
