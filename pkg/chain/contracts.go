package chain

import (
	"math/big"

	"github.com/LiskHQ/op-fault-detector/pkg/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
)

type OracleContract struct {
	contractInstance *bindings.L2OutputOracle
	log              log.Logger
}

func SetContractInstance(contractInstance *bindings.L2OutputOracle, log log.Logger) (*OracleContract, error) {
	return &OracleContract{
		contractInstance: contractInstance,
		log:              log,
	}, nil
}

func GetL1OracleContractAddressByChainID(chainID uint64) string {
	ContractAddresses := GetContractAddressesByChainID(chainID)
	address := ContractAddresses["l1"].l2OutputOracle
	return address
}

func OracleContractInstance(client *ethclient.Client, chainID uint64, log log.Logger) (*bindings.L2OutputOracle, error) {
	oracleContractAddress := GetL1OracleContractAddressByChainID(chainID)

	contract, err := bindings.NewL2OutputOracle(common.HexToAddress(oracleContractAddress), client)

	if err != nil {
		return nil, err
	}

	return contract, nil
}

// CreateContractInstance return [OracleContract] with contract instance.
func CreateContractInstance(url string, chainID uint64, logger log.Logger) (*OracleContract, error) {
	client, err := ethclient.Dial(url)

	if err != nil {
		logger.Errorf("Error occurred while connecting %w", err)
	}

	contractInstance, err := OracleContractInstance(client, chainID, logger)

	if err != nil {
		return nil, err
	}

	return SetContractInstance(contractInstance, logger)
}

func (oc *OracleContract) GetNextOutputIndex() *big.Int {
	nextOutputIndex, err := oc.contractInstance.NextOutputIndex(&bind.CallOpts{})

	if err != nil {
		oc.log.Errorf("Error occurred while retrieving next output index %w", err)
	}

	return nextOutputIndex
}

func (oc *OracleContract) GetL2Output(index *big.Int) bindings.TypesOutputProposal {
	l2Output, err := oc.contractInstance.GetL2Output(&bind.CallOpts{}, index)

	if err != nil {
		oc.log.Errorf("Error occurred while retrieving L2 outout %w", err)
	}

	return l2Output
}
