package encoding

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMustConvertBigIntToUint64(t *testing.T) {
	assert := assert.New(t)

	bigIntInput := big.NewInt(1000)
	const expectedOutput uint64 = 1000
	result := MustConvertBigIntToUint64(bigIntInput)
	assert.Equal(expectedOutput, result)
}

func TestMustConvertUint64ToBigInt(t *testing.T) {
	assert := assert.New(t)

	const uint64Output uint64 = 1000
	expectedOutput := big.NewInt(1000)
	result := MustConvertUint64ToBigInt(uint64Output)
	assert.Equal(expectedOutput, result)
}
