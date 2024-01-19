package chain

import (
	"math/big"

	"github.com/LiskHQ/op-fault-detector/pkg/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
)

// OracleAccessor binds oracle contract to an instance for querying data.
type OracleAccessor struct {
	contractInstance *bindings.L2OutputOracle
	log              log.Logger
}

// getL1OracleContractAddressByChainID returns L1 oracle contract address by chainID.
func getL1OracleContractAddressByChainID(chainID uint64) string {
	ContractAddresses := GetContractAddressesByChainID(chainID)
	address := ContractAddresses["l1"].l2OutputOracle
	return address
}

// NewOracleContract returns [OracleAccessor] with contract instance.
func NewOracleContract(url string, chainID uint64, logger log.Logger) (*OracleAccessor, error) {
	client, err := ethclient.Dial(url)

	if err != nil {
		return nil, err
	}

	oracleContractAddress := getL1OracleContractAddressByChainID(chainID)
	contractInstance, err := bindings.NewL2OutputOracle(common.HexToAddress(oracleContractAddress), client)

	if err != nil {
		return nil, err
	}

	return &OracleAccessor{
		contractInstance: contractInstance,
		log:              logger,
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
