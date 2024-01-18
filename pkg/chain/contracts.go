package chain

import (
	"math/big"

	"github.com/LiskHQ/op-fault-detector/pkg/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
)

// OracleContract binds oracle contract to an instance for querying data.
type OracleContract struct {
	contractInstance *bindings.L2OutputOracle
	log              log.Logger
}

// GetL1OracleContractAddressByChainID returns L1 oracle contract address by chainID.
func GetL1OracleContractAddressByChainID(chainID uint64) string {
	ContractAddresses := GetContractAddressesByChainID(chainID)
	address := ContractAddresses["l1"].l2OutputOracle
	return address
}

// CreateOracleContractInstance returns [OracleContract] with contract instance.
func CreateOracleContractInstance(url string, chainID uint64, logger log.Logger) (*OracleContract, error) {
	client, err := ethclient.Dial(url)

	if err != nil {
		return nil, err
	}

	oracleContractAddress := GetL1OracleContractAddressByChainID(chainID)
	contractInstance, err := bindings.NewL2OutputOracle(common.HexToAddress(oracleContractAddress), client)

	if err != nil {
		return nil, err
	}

	return &OracleContract{
		contractInstance: contractInstance,
		log:              logger,
	}, nil
}

// GetNextOutputIndex returns index of next output to be proposed.
func (oc *OracleContract) GetNextOutputIndex() *big.Int {
	nextOutputIndex, err := oc.contractInstance.NextOutputIndex(&bind.CallOpts{})

	if err != nil {
		oc.log.Errorf("Error occurred while retrieving next output index %w", err)
	}

	return nextOutputIndex
}

// GetL2Output returns L2 output at given index.
func (oc *OracleContract) GetL2Output(index *big.Int) bindings.TypesOutputProposal {
	l2Output, err := oc.contractInstance.GetL2Output(&bind.CallOpts{}, index)

	if err != nil {
		oc.log.Errorf("Error occurred while retrieving L2 outout %w", err)
	}

	return l2Output
}
