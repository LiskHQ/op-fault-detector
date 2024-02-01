package encoding

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
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

func TestComputeL2OutputRoot(t *testing.T) {
	assert := assert.New(t)

	stateRoot := common.HexToHash("0x80f629c32f1c1f00f6ba69825447834fd38ab2cbcacad1afd85735dfcaa195e9")
	storageRoot := common.HexToHash("0xe90f18fe430dfa10aa5f3d052170e219e051c9c954f79bb034e757ab98a8f9d7")
	blockHash := common.HexToHash("0xe77ab8d9935e2e7e25a2169760668a1e45208cf9afe117dc19b91d35bd4a1aa6")
	expectedOutputRoot := "0xca38ae831225597779d84494aa73ccb91e7d497576ed853bf1e9273962dd1884"
	result := ComputeL2OutputRoot(stateRoot, storageRoot, blockHash)
	assert.Equal(expectedOutputRoot, result)
}
