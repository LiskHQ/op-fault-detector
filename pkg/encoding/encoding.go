// Package encoding implements all the utils/helpers for encoding incoming/outgoing RPC requests
package encoding

import (
	"math/big"

	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// MustConvertBigIntToUint64 converts big integer to integer.
func MustConvertBigIntToUint64(value *big.Int) uint64 {
	valueToString, ok := new(big.Int).SetString(value.String(), 10)
	if !ok {
		panic("error while converting big integer to integer")
	}
	return valueToString.Uint64()
}

// MustConvertUint64ToBigInt converts integer to big integer.
func MustConvertUint64ToBigInt(value uint64) *big.Int {
	return new(big.Int).SetUint64(value)
}

// ComputeL2OutputRoot computes L2 output root.
func ComputeL2OutputRoot(stateRoot common.Hash, storageHash common.Hash, outputBlockHash common.Hash) common.Hash {
	l2Output := eth.OutputV0{
		StateRoot:                eth.Bytes32(stateRoot),
		MessagePasserStorageRoot: eth.Bytes32(storageHash),
		BlockHash:                outputBlockHash,
	}

	outputRoot := crypto.Keccak256Hash(l2Output.Marshal())
	return outputRoot
}
