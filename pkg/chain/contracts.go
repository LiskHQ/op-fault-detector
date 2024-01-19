package chain

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
)

// OracleAccessor binds oracle contract to an instance for querying data.
type OracleAccessor struct {
	contractInstance *bindings.L2OutputOracle
}

type ConfigOptions struct {
	L1RPCEndpoint                 string
	ChainID                       uint64
	L2OutputOracleContractAddress string
}

// getL1OracleContractAddressByChainID returns L1 oracle contract address by chainID.
func getL1OracleContractAddressByChainID(chainID uint64) (string, error) {
	ContractAddresses, err := GetContractAddressesByChainID(chainID)
	if err != nil {
		return "", err
	}

	address := ContractAddresses.l2OutputOracle
	return address, nil
}

// NewOracleContract returns [OracleAccessor] with contract instance.
func NewOracleContract(ctx context.Context, opts ConfigOptions) (*OracleAccessor, error) {
	client, err := ethclient.DialContext(ctx, opts.L1RPCEndpoint)
	if err != nil {
		return nil, err
	}

	oracleContractAddress, err := getL1OracleContractAddressByChainID(opts.ChainID)

	// Verify if oracle contract address is available in the chain constants
	// If not available, use l2OutputContractAddress from the config options
	if err != nil {
		if len(opts.L2OutputOracleContractAddress) > 0 {
			oracleContractAddress = opts.L2OutputOracleContractAddress
		} else {
			return nil, fmt.Errorf("L2 output oracle contract address is not available")
		}
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
func (oc *OracleAccessor) GetL2Output(index *big.Int) (bindings.TypesOutputProposal, error) {
	return oc.contractInstance.GetL2Output(&bind.CallOpts{}, index)
}
