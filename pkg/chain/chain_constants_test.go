package chain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetContractAddressesByChainID_AvailableChainID(t *testing.T) {
	const availableChainID uint64 = 10
	contractAddresses := GetContractAddressesByChainID(availableChainID)
	contractAddressesExpected := map[string]L1Contracts{
		"l1": {
			stateCommitmentChain: "0xBe5dAb4A2e9cd0F27300dB4aB94BeE3A233AEB19",
			optimismPortal:       "0xbEb5Fc579115071764c7423A4f12eDde41f106Ed",
			l2OutputOracle:       "0xdfe97868233d1aa22e815a266982f2cf17685a27",
		}}

	assert.Equal(t, contractAddressesExpected, contractAddresses)
}

func TestGetContractAddressesByChainID_UnavailableChainID(t *testing.T) {
	const unavailableChainID uint64 = 5
	contractAddresses := GetContractAddressesByChainID(unavailableChainID)

	assert.Equal(t, 0, len(contractAddresses))
}
