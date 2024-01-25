// Package encoding implements all the utils/helpers for encoding incoming/outgoing RPC requests
package encoding

import "math/big"

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
