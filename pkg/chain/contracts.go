package chain

import (
	"math/big"

	"github.com/LiskHQ/op-fault-detector/pkg/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
)

func GetL1OracleContractAddressByChainID(chainID uint64) string {
	ContractAddresses := GetContractAddresses(chainID)
	address := ContractAddresses["l1"].l2OutputOracle
	return address
}

// TODO: Create oracle Struct with required functions
func OracleContractInstance(client *ethclient.Client, chainID uint64, log log.Logger) (*bindings.L2OutputOracle, error) {
	oracleContractAddress := GetL1OracleContractAddressByChainID(chainID)

	contract, err := bindings.NewL2OutputOracle(common.HexToAddress(oracleContractAddress), client)

	if err != nil {
		return nil, err
	}

	return contract, nil
}

// TODO: Use EthClientInterface
func CreateContractInstance(url string, chainID uint64, logger log.Logger) *bindings.L2OutputOracle {
	client, err := ethclient.Dial(url)

	if err != nil {
		logger.Errorf("Error occurred while connecting %w", err)
	}

	contract, err := OracleContractInstance(client, chainID, logger)

	if err != nil {
		logger.Errorf("Error occurred while creating contract instance %w", err)
	}

	return contract
}

func GetNextOutputIndex(contractInstance *bindings.L2OutputOracle, log log.Logger) *big.Int {
	nextOutputIndex, err := contractInstance.NextOutputIndex(&bind.CallOpts{})

	if err != nil {
		log.Errorf("Error occurred while retrieving next output index %w", err)
	}

	return nextOutputIndex
}

func GetL2Output(contractInstance *bindings.L2OutputOracle, index *big.Int, log log.Logger) bindings.TypesOutputProposal {
	l2Output, err := contractInstance.GetL2Output(&bind.CallOpts{}, index)

	if err != nil {
		log.Errorf("Error occurred while retrieving L2 outout %w", err)
	}

	return l2Output
}
