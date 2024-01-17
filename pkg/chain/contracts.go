package chain

import (
	"math/big"

	"github.com/LiskHQ/op-fault-detector/pkg/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
)

func GetOracleAddressbyChainID(l2ChainID int) string {
	ContractAddresses := getContractAddresses()
	address := ContractAddresses[l2ChainID]["l1"].L2OutputOracle
	return address
}

// TODO: Create oracle Struct with required functions
func OracleContractInstance(client *ethclient.Client, l2ChainID int, log log.Logger) (*bindings.L2OutputOracle, error) {
	oracleContractAddress := GetOracleAddressbyChainID(l2ChainID)

	contract, err := bindings.NewL2OutputOracle(common.HexToAddress(oracleContractAddress), client)

	if err != nil {
		return nil, err
	}

	return contract, nil
}

// TODO: Use EthClientInterface
func CreateContractInstance(url string, l2ChainID int, logger log.Logger) *bindings.L2OutputOracle {
	client, err := ethclient.Dial(url)

	if err != nil {
		logger.Errorf("Error occurred while connecting %w", err)
	}

	contract, err := OracleContractInstance(client, l2ChainID, logger)

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
