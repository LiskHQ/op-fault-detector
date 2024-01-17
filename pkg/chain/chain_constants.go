package chain

import (
	"reflect"
)

type DefaultL2ContractAddress struct {
	L2CrossDomainMessenger  string
	L2ToL1MessagePasser     string
	L2StandardBridge        string
	OVM_L1BlockNumber       string
	OVM_L2ToL1MessagePasser string
	OVM_DeployerWhitelist   string
	OVM_ETH                 string
	OVM_GasPriceOracle      string
	OVM_SequencerFeeVault   string
	WETH                    string
	BedrockMessagePasser    string
}

type NetworkType struct {
	mainnet string
	goerli  string
	sepolia string
}

type L2ChainIDs struct {
	OPTIMISM                       int
	OPTIMISM_GOERLI                int
	OPTIMISM_HARDHAT_LOCAL         int
	OPTIMISM_HARDHAT_DEVNET        int
	OPTIMISM_BEDROCK_LOCAL_DEVNET  int
	OPTIMISM_BEDROCK_ALPHA_TESTNET int
	BASE_GOERLI                    int
	ZORA_GOERLI                    int
	ZORA_MAINNET                   int
	OPTIMISM_SEPOLIA               int
	BASE_SEPOLIA                   int
	BASE_MAINNET                   int
}

var L2_CHAIN_ID = L2ChainIDs{
	OPTIMISM:                       10,
	OPTIMISM_GOERLI:                420,
	OPTIMISM_SEPOLIA:               11155420,
	OPTIMISM_HARDHAT_LOCAL:         31337,
	OPTIMISM_HARDHAT_DEVNET:        17,
	OPTIMISM_BEDROCK_ALPHA_TESTNET: 28528,
	BASE_GOERLI:                    84531,
	BASE_SEPOLIA:                   84532,
	BASE_MAINNET:                   8453,
	ZORA_GOERLI:                    999,
	ZORA_MAINNET:                   7777777,
}

var L2_OUTPUT_ORACLE_ADDRESSES = NetworkType{
	mainnet: "0xdfe97868233d1aa22e815a266982f2cf17685a27",
	goerli:  "0xE6Dfba0953616Bacab0c9A8ecb3a9BBa77FC15c0",
	sepolia: "0x90E9c4f8a994a250F6aEfd61CAFb4F2e895D458F",
}

var L1_CROSS_DOMAIN_MESSENGER = NetworkType{
	mainnet: "0x25ace71c97B33Cc4729CF772ae268934F7ab5fA1",
	goerli:  "0x5086d1eEF304eb5284A0f6720f79403b4e9bE294",
	sepolia: "0x58Cc85b8D04EA49cC6DBd3CbFFd00B4B8D6cb3ef",
}

var STATE_COMMITMENT_CHAIN = NetworkType{
	mainnet: "0xBe5dAb4A2e9cd0F27300dB4aB94BeE3A233AEB19",
	goerli:  "0x9c945aC97Baf48cB784AbBB61399beB71aF7A378",
	sepolia: "0x0000000000000000000000000000000000000000",
}

var OPTIMISM_PORTAL_ADDRESS = NetworkType{
	mainnet: "0xbEb5Fc579115071764c7423A4f12eDde41f106Ed",
	goerli:  "0x5b47E1A08Ea6d985D6649300584e6722Ec4B1383",
	sepolia: "0x16Fc5058F25648194471939df75CF27A2fdC48BC",
}

type L1Contracts struct {
	L1CrossDomainMessenger string
	StateCommitmentChain   string
	OptimismPortal         string
	L2OutputOracle         string
}

type ContractAddresses struct {
	L1Contracts
	DefaultL2ContractAddress
}

var DEFAULT_L2_CONTRACT_ADDRESS = DefaultL2ContractAddress{
	BedrockMessagePasser:    "0x4200000000000000000000000000000000000016",
	L2CrossDomainMessenger:  "0x4200000000000000000000000000000000000007",
	L2StandardBridge:        "0x4200000000000000000000000000000000000010",
	L2ToL1MessagePasser:     "0x4200000000000000000000000000000000000016",
	OVM_DeployerWhitelist:   "0x4200000000000000000000000000000000000002",
	OVM_ETH:                 "0xDeadDeAddeAddEAddeadDEaDDEAdDeaDDeAD0000",
	OVM_GasPriceOracle:      "0x420000000000000000000000000000000000000F",
	OVM_L1BlockNumber:       "0x4200000000000000000000000000000000000013",
	OVM_L2ToL1MessagePasser: "0x4200000000000000000000000000000000000016",
	OVM_SequencerFeeVault:   "0x4200000000000000000000000000000000000011",
	WETH:                    "0x4200000000000000000000000000000000000006",
}

func FilterAddressByNetwork(c NetworkType, network string) string {
	ref := reflect.ValueOf(c)
	f := reflect.Indirect(ref).FieldByName(network)
	return f.String()
}

func getL1ContractsByNetworkName(network string) L1Contracts {
	L1Contracts := L1Contracts{
		L1CrossDomainMessenger: FilterAddressByNetwork(L1_CROSS_DOMAIN_MESSENGER, network),
		StateCommitmentChain:   FilterAddressByNetwork(STATE_COMMITMENT_CHAIN, network),
		OptimismPortal:         FilterAddressByNetwork(OPTIMISM_PORTAL_ADDRESS, network),
		L2OutputOracle:         FilterAddressByNetwork(L2_OUTPUT_ORACLE_ADDRESSES, network),
	}

	return L1Contracts
}

func getContractAddresses() map[int]map[string]L1Contracts {
	CONTRACT_ADDRESSES := map[int]map[string]L1Contracts{
		L2_CHAIN_ID.OPTIMISM: {
			"l1": getL1ContractsByNetworkName("mainnet"),
			// "l2": DEFAULT_L2_CONTRACT_ADDRESS,
		},
		L2_CHAIN_ID.OPTIMISM_GOERLI: {
			"l1": getL1ContractsByNetworkName("goerli"),
			// "l2": DEFAULT_L2_CONTRACT_ADDRESS,
		},
		L2_CHAIN_ID.OPTIMISM_SEPOLIA: {
			"l1": getL1ContractsByNetworkName("sepolia"),
			// "l2": DEFAULT_L2_CONTRACT_ADDRESS,
		},
		L2_CHAIN_ID.OPTIMISM_HARDHAT_LOCAL: {
			"l1": {
				L1CrossDomainMessenger: "0x8A791620dd6260079BF849Dc5567aDC3F2FdC318",
				StateCommitmentChain:   "0xDc64a140Aa3E981100a9becA4E685f962f0cF6C9",
				OptimismPortal:         "0x0000000000000000000000000000000000000000",
				L2OutputOracle:         "0x0000000000000000000000000000000000000000",
			},
			// "l2": DEFAULT_L2_CONTRACT_ADDRESS,
		},
		L2_CHAIN_ID.OPTIMISM_HARDHAT_DEVNET: {
			"l1": {
				L1CrossDomainMessenger: "0x8A791620dd6260079BF849Dc5567aDC3F2FdC318",
				StateCommitmentChain:   "0xDc64a140Aa3E981100a9becA4E685f962f0cF6C9",
				OptimismPortal:         "0x0000000000000000000000000000000000000000",
				L2OutputOracle:         "0x0000000000000000000000000000000000000000",
			},
			// "l2": DEFAULT_L2_CONTRACT_ADDRESS,
		},
		L2_CHAIN_ID.OPTIMISM_BEDROCK_ALPHA_TESTNET: {
			"l1": {
				L1CrossDomainMessenger: "0x838a6DC4E37CA45D4Ef05bb776bf05eEf50798De",
				StateCommitmentChain:   "0x0000000000000000000000000000000000000000",
				OptimismPortal:         "0xA581Ca3353DB73115C4625FFC7aDF5dB379434A8",
				L2OutputOracle:         "0x3A234299a14De50027eA65dCdf1c0DaC729e04A6",
			},
			// "l2": DEFAULT_L2_CONTRACT_ADDRESS,
		},
		L2_CHAIN_ID.BASE_GOERLI: {
			"l1": {
				L1CrossDomainMessenger: "0x8e5693140eA606bcEB98761d9beB1BC87383706D",
				StateCommitmentChain:   "0x0000000000000000000000000000000000000000",
				OptimismPortal:         "0xe93c8cD0D409341205A592f8c4Ac1A5fe5585cfA",
				L2OutputOracle:         "0x2A35891ff30313CcFa6CE88dcf3858bb075A2298",
			},
			// "l2": DEFAULT_L2_CONTRACT_ADDRESS,
		},
		L2_CHAIN_ID.BASE_SEPOLIA: {
			"l1": {
				L1CrossDomainMessenger: "0xC34855F4De64F1840e5686e64278da901e261f20",
				StateCommitmentChain:   "0x0000000000000000000000000000000000000000",
				OptimismPortal:         "0x49f53e41452C74589E85cA1677426Ba426459e85",
				L2OutputOracle:         "0x84457ca9D0163FbC4bbfe4Dfbb20ba46e48DF254",
			},
			// "l2": DEFAULT_L2_CONTRACT_ADDRESS,
		},
		L2_CHAIN_ID.BASE_MAINNET: {
			"l1": {
				L1CrossDomainMessenger: "0x866E82a600A1414e583f7F13623F1aC5d58b0Afa",
				StateCommitmentChain:   "0x0000000000000000000000000000000000000000",
				OptimismPortal:         "0x49048044D57e1C92A77f79988d21Fa8fAF74E97e",
				L2OutputOracle:         "0x56315b90c40730925ec5485cf004d835058518A0",
			},
			// "l2": DEFAULT_L2_CONTRACT_ADDRESS,
		},
		L2_CHAIN_ID.ZORA_GOERLI: {
			"l1": {
				L1CrossDomainMessenger: "0xD87342e16352D33170557A7dA1e5fB966a60FafC",
				StateCommitmentChain:   "0x0000000000000000000000000000000000000000",
				OptimismPortal:         "0xDb9F51790365e7dc196e7D072728df39Be958ACe",
				L2OutputOracle:         "0xdD292C9eEd00f6A32Ff5245d0BCd7f2a15f24e00",
			},
			// "l2": DEFAULT_L2_CONTRACT_ADDRESS,
		},
		L2_CHAIN_ID.ZORA_MAINNET: {
			"l1": {
				L1CrossDomainMessenger: "0xdC40a14d9abd6F410226f1E6de71aE03441ca506",
				StateCommitmentChain:   "0x0000000000000000000000000000000000000000",
				OptimismPortal:         "0x1a0ad011913A150f69f6A19DF447A0CfD9551054",
				L2OutputOracle:         "0x9E6204F750cD866b299594e2aC9eA824E2e5f95c",
			},
			// "l2": DEFAULT_L2_CONTRACT_ADDRESS,
		},
	}

	return CONTRACT_ADDRESSES
}
