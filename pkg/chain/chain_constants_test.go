package chain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetContractAddressesByChainID(t *testing.T) {
	assert := assert.New(t)

	const availableChainID uint64 = 10
	contractAddresses, areAddressesAvailable := GetContractAddressesByChainID(availableChainID)
	contractAddressesExpected := Contracts{
		stateCommitmentChain: "0xBe5dAb4A2e9cd0F27300dB4aB94BeE3A233AEB19",
		optimismPortal:       "0xbEb5Fc579115071764c7423A4f12eDde41f106Ed",
		l2OutputOracle:       "0xdfe97868233d1aa22e815a266982f2cf17685a27",
		networkType:          "L1",
	}
	assert.Equal(true, areAddressesAvailable)
	assert.Equal(contractAddressesExpected, contractAddresses)

	const unavailableChainID uint64 = 5
	contractAddresses, areAddressesAvailable = GetContractAddressesByChainID(unavailableChainID)
	assert.Equal(false, areAddressesAvailable)
	assert.Equal(0, len(contractAddresses.l2OutputOracle))
}
