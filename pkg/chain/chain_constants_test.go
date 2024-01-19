package chain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetContractAddressesByChainID_AvailableChainID(t *testing.T) {
	const availableChainID uint64 = 10
	contractAddresses, err := GetContractAddressesByChainID(availableChainID)

	contractAddressesExpected := Contracts{
		stateCommitmentChain: "0xBe5dAb4A2e9cd0F27300dB4aB94BeE3A233AEB19",
		optimismPortal:       "0xbEb5Fc579115071764c7423A4f12eDde41f106Ed",
		l2OutputOracle:       "0xdfe97868233d1aa22e815a266982f2cf17685a27",
		networkType:          "L1",
	}

	assert.NoError(t, err)
	assert.Equal(t, contractAddressesExpected, contractAddresses)
}

func TestGetContractAddressesByChainID_UnavailableChainID(t *testing.T) {
	const unavailableChainID uint64 = 5
	contractAddresses, err := GetContractAddressesByChainID(unavailableChainID)

	assert.Error(t, err)
	assert.Equal(t, 0, len(contractAddresses.l2OutputOracle))
}
