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
	Network_Prater  Network = "prater"
	Network_Devnet  Network = "devnet"
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
	ExecutionClient_Obs_Infura ExecutionClient = "infura"
	ExecutionClient_Obs_Pocket ExecutionClient = "pocket"
)

// Enum to describe the Consensus client options
const (
	ConsensusClient_Unknown    ConsensusClient = ""
	ConsensusClient_Lighthouse ConsensusClient = "lighthouse"
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
	MevRelayID_Blocknative        MevRelayID = "blocknative"
	MevRelayID_Eden               MevRelayID = "eden"
	MevRelayID_Ultrasound         MevRelayID = "ultrasound"
)

// Enum to describe MEV-Boost relay selection mode
const (
	MevSelectionMode_Profile MevSelectionMode = "profile"
	MevSelectionMode_Relay   MevSelectionMode = "relay"
)

type Config interface {
	GetConfigTitle() string
	GetParameters() []*Parameter
}

// Interface for common Consensus configurations
type ConsensusConfig interface {
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
	ID            MevRelayID
	Name          string
	Description   string
	Urls          map[Network]string
	Regulated     bool
	NoSandwiching bool
}
