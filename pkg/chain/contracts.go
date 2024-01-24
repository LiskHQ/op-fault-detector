package chain

import (
	"context"
	"fmt"
	"math/big"

	"github.com/LiskHQ/op-fault-detector/pkg/encoding"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
)

// L2Output is the output of GetL2Output.
type L2Output struct {
	OutputRoot    string
	L1Timestamp   uint64
	L2BlockNumber uint64
	L2OutputIndex uint64
}

// OracleAccessor binds oracle contract to an instance for querying data.
type OracleAccessor struct {
	contractInstance *bindings.L2OutputOracle
}

// ConfigOptions are the options required to interact with the oracle contract.
type ConfigOptions struct {
	L1RPCEndpoint                 string
	ChainID                       uint64
	L2OutputOracleContractAddress string
}

func getL1OracleContractAddressByChainID(chainID uint64) (string, bool) {
	cAddr, areAddressesAvailable := GetContractAddressesByChainID(chainID)
	if !areAddressesAvailable {
		return "", false
	}

	return cAddr.l2OutputOracle, true
}

// NewOracleAccessor returns [OracleAccessor] with contract instance.
func NewOracleAccessor(ctx context.Context, opts *ConfigOptions) (*OracleAccessor, error) {
	client, err := ethclient.DialContext(ctx, opts.L1RPCEndpoint)
	if err != nil {
		return nil, err
	}

	oracleContractAddress, isAddressExists := getL1OracleContractAddressByChainID(opts.ChainID)

	// Verify if oracle contract address is available in the chain constants
	// If not available, use l2OutputContractAddress from the config options
	if !isAddressExists {
		if len(opts.L2OutputOracleContractAddress) == 0 {
			return nil, fmt.Errorf("L2OutputOracleContractAddress is not available")
		}
		oracleContractAddress = opts.L2OutputOracleContractAddress
	}

	oracleContractInstance, err := bindings.NewL2OutputOracle(common.HexToAddress(oracleContractAddress), client)

	if err != nil {
		return nil, err
	}

	return &OracleAccessor{
		contractInstance: oracleContractInstance,
	}, nil
}

// GetNextOutputIndex returns index of next output to be proposed.
func (oc *OracleAccessor) GetNextOutputIndex() (*big.Int, error) {
	return oc.contractInstance.NextOutputIndex(&bind.CallOpts{})
}

// GetL2Output returns L2 output at given index.
func (oc *OracleAccessor) GetL2Output(index *big.Int) (L2Output, error) {
	l2Output, err := oc.contractInstance.GetL2Output(&bind.CallOpts{}, index)
	if err != nil {
		return L2Output{}, err
	}

	return L2Output{
		OutputRoot:    "0x" + common.Bytes2Hex(l2Output.OutputRoot[:]),
		L1Timestamp:   encoding.ConvertBigIntToUint64(l2Output.Timestamp),
		L2BlockNumber: encoding.ConvertBigIntToUint64(l2Output.L2BlockNumber),
		L2OutputIndex: encoding.ConvertBigIntToUint64(index),
	}, nil
}
