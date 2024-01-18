package chain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetL1OracleContractAddressByChainID(t *testing.T) {
	const availableChainID uint64 = 10
	var contractAddresses = GetL1OracleContractAddressByChainID(availableChainID)
	var oracleContractAddressesExpected = "0xdfe97868233d1aa22e815a266982f2cf17685a27"

	assert.Equal(t, oracleContractAddressesExpected, contractAddresses)

	const unavailableChainID uint64 = 5
	contractAddresses = GetL1OracleContractAddressByChainID(unavailableChainID)
	assert.Equal(t, 0, len(contractAddresses))
}
