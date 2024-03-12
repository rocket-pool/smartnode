package config

type ContainerID string
type Network string
type Mode string
type ParameterType string
type ExecutionClient string
type ConsensusClient string
type RewardsMode string
type MevRelayID string
type MevSelectionMode string
type NimbusPruningMode string

// Enum to describe which container(s) a parameter impacts, so the Smartnode knows which
// ones to restart upon a settings change
const (
	ContainerID_Unknown    ContainerID = ""
	ContainerID_Api        ContainerID = "api"
	ContainerID_Node       ContainerID = "node"
	ContainerID_Watchtower ContainerID = "watchtower"
	ContainerID_Eth1       ContainerID = "eth1"
	ContainerID_Eth2       ContainerID = "eth2"
	ContainerID_Validator  ContainerID = "validator"
	ContainerID_Grafana    ContainerID = "grafana"
	ContainerID_Prometheus ContainerID = "prometheus"
	ContainerID_Exporter   ContainerID = "exporter"
	ContainerID_MevBoost   ContainerID = "mev-boost"
)

// Enum to describe which network the system is on
const (
	Network_Unknown Network = ""
	Network_All     Network = "all"
	Network_Mainnet Network = "mainnet"
	Network_Devnet  Network = "devnet"
	Network_Holesky Network = "holesky"
)

// Enum to describe the mode for a client - local (Docker Mode) or external (Hybrid Mode)
const (
	Mode_Unknown  Mode = ""
	Mode_Local    Mode = "local"
	Mode_External Mode = "external"
)

// Enum to describe which data type a parameter's value will have, which
// informs the corresponding UI element and value validation
const (
	ParameterType_Unknown ParameterType = ""
	ParameterType_Int     ParameterType = "int"
	ParameterType_Uint16  ParameterType = "uint16"
	ParameterType_Uint    ParameterType = "uint"
	ParameterType_String  ParameterType = "string"
	ParameterType_Bool    ParameterType = "bool"
	ParameterType_Choice  ParameterType = "choice"
	ParameterType_Float   ParameterType = "float"
)

// Enum to describe the Execution client options
const (
	ExecutionClient_Unknown    ExecutionClient = ""
	ExecutionClient_Geth       ExecutionClient = "geth"
	ExecutionClient_Nethermind ExecutionClient = "nethermind"
	ExecutionClient_Besu       ExecutionClient = "besu"
	ExecutionClient_Reth       ExecutionClient = "reth"
)

// Enum to describe the Consensus client options
const (
	ConsensusClient_Unknown    ConsensusClient = ""
	ConsensusClient_Lighthouse ConsensusClient = "lighthouse"
	ConsensusClient_Lodestar   ConsensusClient = "lodestar"
	ConsensusClient_Nimbus     ConsensusClient = "nimbus"
	ConsensusClient_Prysm      ConsensusClient = "prysm"
	ConsensusClient_Teku       ConsensusClient = "teku"
)

// Enum to describe the rewards tree acquisition modes
const (
	RewardsMode_Unknown  RewardsMode = ""
	RewardsMode_Download RewardsMode = "download"
	RewardsMode_Generate RewardsMode = "generate"
)

// Enum to identify MEV-boost relays
const (
	MevRelayID_Unknown            MevRelayID = ""
	MevRelayID_Flashbots          MevRelayID = "flashbots"
	MevRelayID_BloxrouteEthical   MevRelayID = "bloxrouteEthical"
	MevRelayID_BloxrouteMaxProfit MevRelayID = "bloxrouteMaxProfit"
	MevRelayID_BloxrouteRegulated MevRelayID = "bloxrouteRegulated"
	MevRelayID_Eden               MevRelayID = "eden"
	MevRelayID_Ultrasound         MevRelayID = "ultrasound"
	MevRelayID_Aestus             MevRelayID = "aestus"
)

// Enum to describe MEV-Boost relay selection mode
const (
	MevSelectionMode_Profile MevSelectionMode = "profile"
	MevSelectionMode_Relay   MevSelectionMode = "relay"
)

// Enum to describe Nimbus pruning modes
const (
	NimbusPruningMode_Archive NimbusPruningMode = "archive"
	NimbusPruningMode_Prune   NimbusPruningMode = "prune"
)

type Config interface {
	GetConfigTitle() string
	GetParameters() []*Parameter
}

// Interface for common Consensus configurations
type ConsensusConfig interface {
	GetBeaconNodeImage() string
	GetValidatorImage() string
	GetName() string
}

// Interface for Local Consensus configurations
type LocalConsensusConfig interface {
	GetUnsupportedCommonParams() []string
}

// Interface for External Consensus configurations
type ExternalConsensusConfig interface {
	GetApiUrl() string
}

// A setting that has changed
type ChangedSetting struct {
	Name               string
	OldValue           string
	NewValue           string
	AffectedContainers map[ContainerID]bool
}

// A MEV relay
type MevRelay struct {
	ID          MevRelayID
	Name        string
	Description string
	Urls        map[Network]string
	Regulated   bool
}

// Describes a network that the smartnode system can use.
type NetworkInfo struct {
	// A unique name for the network
	Name string `yaml:"name"`
	// A human-readable label for the network
	Label string `yaml:"label"`
	// A human-readable description of the network
	Description string `yaml:"description"`
	// The URL to provide the user so they can follow pending transactions
	TxWatchUrl string `yaml:"txWatchUrl"`
	// The URL to use for staking rETH
	StakeUrl string `yaml:"stakeUrl"`
	// Execution chain ID
	ChainID uint64 `yaml:"chainID"`
	// The Snapshot API domain
	SnapshotApiDomain string `yaml:"snapshotApiDomain"`
	// The FlashBots Protect RPC endpoint
	FlashbotsProtectUrl string `yaml:"flashbotsProtectUrl"`
	// Indicates if we support mevboost on the network
	IsMevBoostSupported bool `yaml:"isMevBoostSupported"`
	// The set of addresses we need for the network.
	Addresses NetworkAddresses `yaml:"addresses"`
}

// Describes the set of addresses for a network.
type NetworkAddresses struct {
	// The contract address of RocketStorage
	Storage string `yaml:"storage"`
	// The contract address of the RPL token
	RplToken string `yaml:"rplToken"`
	// The contract address of the RPL faucet
	RplFaucet string `yaml:"rplFaucet"`
	// The contract address for Snapshot delegation
	SnapshotDelegation string `yaml:"snapshotDelegation"`
	// The contract address of rETH
	Reth string `yaml:"reth"`
	// The contract address of rocketRewardsPool from v1.0.0
	V1_0_0_RewardsPool string `yaml:"v1_0_0_RewardsPool"`
	// The contract address of rocketClaimNode from v1.0.0
	V1_0_0_ClaimNode string `yaml:"v1_0_0_ClaimNode"`
	// The contract address of rocketClaimTrustedNode from v1.0.0
	V1_0_0_ClaimTrustedNode string `yaml:"v1_0_0_ClaimTrustedNode"`
	// The contract address of rocketMinipoolManager from v1.0.0
	V1_0_0_MinipoolManager string `yaml:"v1_0_0_MinipoolManager"`
	// The contract address of rocketNetworkPrices from v1.1.0
	V1_1_0_NetworkPrices string `yaml:"v1_1_0_NetworkPrices"`
	// The contract address of rocketNodeStaking from v1.1.0
	V1_1_0_NodeStaking string `yaml:"v1_1_0_NodeStaking"`
	// The contract address of rocketNodeDeposit from v1.1.0
	V1_1_0_NodeDeposit string `yaml:"v1_1_0_NodeDeposit"`
	// The contract address of rocketMinipoolQueue from v1.1.0
	V1_1_0_MinipoolQueue string `yaml:"v1_1_0_MinipoolQueue"`
	// The contract address of rocketMinipoolFactory from v1.1.0
	V1_1_0_MinipoolFactory string `yaml:"v1_1_0_MinipoolFactory"`
	// Addresses for RocketRewardsPool that have been upgraded during development
	PreviousRewardsPools []string `yaml:"previousRewardsPools"`
	// The RocketOvmPriceMessenger Optimism address for each network
	OptimismPriceMessenger string `yaml:"optimismPriceMessenger"`
	// The RocketPolygonPriceMessenger Polygon address for each network
	PolygonPriceMessenger string `yaml:"polygonPriceMessenger"`
	// The RocketArbitumPriceMessenger Arbitrum address for each network
	ArbitrumPriceMessenger string `yaml:"arbitrumPriceMessenger"`
	// The RocketArbitumPriceMessengerV2 Arbitrum address for each network
	ArbitrumPriceMessengerV2 string `yaml:"arbitrumPriceMessengerV2"`
	// The RocketZkSyncPriceMessenger zkSyncEra address for each network
	ZkSyncEraPriceMessenger string `yaml:"zkSyncEraPriceMessenger"`
	// The RocketBasePriceMessenger Base address for each network
	BasePriceMessenger string `yaml:"basePriceMessenger"`
	// The RocketScrollPriceMessenger Scroll address for each network
	ScrollPriceMessenger string `yaml:"scrollPriceMessenger"`
	// The Scroll L2 message fee estimator address for each network
	ScrollFeeEstimator string `yaml:"scrollFeeEstimator"`
	// The UniswapV3 pool address for each network (used for RPL price TWAP info)
	RplTwapPool string `yaml:"rplTwapPool"`
	// The multicall contract address
	Multicall string `yaml:"multicall"`
	// The BalanceChecker contract address
	Balancebatcher string `yaml:"balancebatcher"`
}
