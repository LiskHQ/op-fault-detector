// Package encoding implements all the utils/helpers for encoding incoming/outgoing RPC requests
package encoding

import "math/big"

// ConvertBigIntToUint64 converts big integer to integer.
func ConvertBigIntToUint64(value *big.Int) uint64 {
	valueToString, _ := new(big.Int).SetString(value.String(), 10)
	return valueToString.Uint64()
}

// ConvertUint64ToBigInt converts integer to big integer.
func ConvertUint64ToBigInt(value uint64) *big.Int {
	return new(big.Int).SetUint64(value)
}
