package chain

// L2ChainIDs manages L2 network chainIDs.
type L2ChainIDs struct {
	optimism                    uint64
	optimismGoerli              uint64
	optimismSepolia             uint64
	optimismHardhatLocal        uint64
	optimismHardhatDevnet       uint64
	optimismBedrockAlphaTestnet uint64
	baseMainnet                 uint64
	baseGoerli                  uint64
	baseSepolia                 uint64
	zoraGoerli                  uint64
	zoraMainnet                 uint64
}

var l2NetworkChainIDs = L2ChainIDs{
	optimism:                    10,
	optimismGoerli:              420,
	optimismSepolia:             11155420,
	optimismHardhatLocal:        31337,
	optimismHardhatDevnet:       17,
	optimismBedrockAlphaTestnet: 28528,
	baseGoerli:                  84531,
	baseSepolia:                 84532,
	baseMainnet:                 8453,
	zoraGoerli:                  999,
	zoraMainnet:                 7777777,
}

// L1Contracts returns L1 network contracts adresses.
type L1Contracts struct {
	stateCommitmentChain string
	optimismPortal       string
	l2OutputOracle       string
}

// GetContractAddressesByChainID returns contract addresses by chainID.
func GetContractAddressesByChainID(chainID uint64) map[string]L1Contracts {
	contractAddresses := map[uint64]map[string]L1Contracts{
		l2NetworkChainIDs.optimism: {
			"l1": {
				stateCommitmentChain: "0xBe5dAb4A2e9cd0F27300dB4aB94BeE3A233AEB19",
				optimismPortal:       "0xbEb5Fc579115071764c7423A4f12eDde41f106Ed",
				l2OutputOracle:       "0xdfe97868233d1aa22e815a266982f2cf17685a27",
			},
		},
		l2NetworkChainIDs.optimismGoerli: {
			"l1": {
				stateCommitmentChain: "0x9c945aC97Baf48cB784AbBB61399beB71aF7A378",
				optimismPortal:       "0x5b47E1A08Ea6d985D6649300584e6722Ec4B1383",
				l2OutputOracle:       "0xE6Dfba0953616Bacab0c9A8ecb3a9BBa77FC15c0",
			},
		},
		l2NetworkChainIDs.optimismSepolia: {
			"l1": {
				stateCommitmentChain: "0x0000000000000000000000000000000000000000",
				optimismPortal:       "0x16Fc5058F25648194471939df75CF27A2fdC48BC",
				l2OutputOracle:       "0x90E9c4f8a994a250F6aEfd61CAFb4F2e895D458F",
			},
		},
		l2NetworkChainIDs.optimismHardhatLocal: {
			"l1": {
				stateCommitmentChain: "0xDc64a140Aa3E981100a9becA4E685f962f0cF6C9",
				optimismPortal:       "0x0000000000000000000000000000000000000000",
				l2OutputOracle:       "0x0000000000000000000000000000000000000000",
			},
		},
		l2NetworkChainIDs.optimismHardhatDevnet: {
			"l1": {
				stateCommitmentChain: "0xDc64a140Aa3E981100a9becA4E685f962f0cF6C9",
				optimismPortal:       "0x0000000000000000000000000000000000000000",
				l2OutputOracle:       "0x0000000000000000000000000000000000000000",
			},
		},
		l2NetworkChainIDs.optimismBedrockAlphaTestnet: {
			"l1": {
				stateCommitmentChain: "0x0000000000000000000000000000000000000000",
				optimismPortal:       "0xA581Ca3353DB73115C4625FFC7aDF5dB379434A8",
				l2OutputOracle:       "0x3A234299a14De50027eA65dCdf1c0DaC729e04A6",
			},
		},
		l2NetworkChainIDs.baseGoerli: {
			"l1": {
				stateCommitmentChain: "0x0000000000000000000000000000000000000000",
				optimismPortal:       "0xe93c8cD0D409341205A592f8c4Ac1A5fe5585cfA",
				l2OutputOracle:       "0x2A35891ff30313CcFa6CE88dcf3858bb075A2298",
			},
		},
		l2NetworkChainIDs.baseSepolia: {
			"l1": {
				stateCommitmentChain: "0x0000000000000000000000000000000000000000",
				optimismPortal:       "0x49f53e41452C74589E85cA1677426Ba426459e85",
				l2OutputOracle:       "0x84457ca9D0163FbC4bbfe4Dfbb20ba46e48DF254",
			},
		},
		l2NetworkChainIDs.baseMainnet: {
			"l1": {
				stateCommitmentChain: "0x0000000000000000000000000000000000000000",
				optimismPortal:       "0x49048044D57e1C92A77f79988d21Fa8fAF74E97e",
				l2OutputOracle:       "0x56315b90c40730925ec5485cf004d835058518A0",
			},
		},
		l2NetworkChainIDs.zoraGoerli: {
			"l1": {
				stateCommitmentChain: "0x0000000000000000000000000000000000000000",
				optimismPortal:       "0xDb9F51790365e7dc196e7D072728df39Be958ACe",
				l2OutputOracle:       "0xdD292C9eEd00f6A32Ff5245d0BCd7f2a15f24e00",
			},
		},
		l2NetworkChainIDs.zoraMainnet: {
			"l1": {
				stateCommitmentChain: "0x0000000000000000000000000000000000000000",
				optimismPortal:       "0x1a0ad011913A150f69f6A19DF447A0CfD9551054",
				l2OutputOracle:       "0x9E6204F750cD866b299594e2aC9eA824E2e5f95c",
			},
		},
	}

	return contractAddresses[chainID]
}
