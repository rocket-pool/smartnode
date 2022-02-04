package config

type ContainerID int
type Network int
type ParameterType int

// Enum to describe which container(s) a parameter impacts, so the Smartnode knows which
// ones to restart upon a settings change
const (
	ContainerID_Unknown ContainerID = iota
	ContainerID_Api
	ContainerID_Node
	ContainerID_Watchtower
	ContainerID_Eth1
	ContainerID_Eth2
	ContainerID_Validator
	ContainerID_Grafana
	ContainerID_Prometheus
	ContainerID_Exporter
)

// Enum to describe which network the system is on
const (
	Network_Unknown Network = iota
	Network_Mainnet
	Network_Prater
)

// Enum to describe which data type a parameter's value will have, which
// informs the corresponding UI element and value validation
const (
	ParameterType_Unknown ParameterType = iota
	ParameterType_Int
	ParameterType_Uint16
	ParameterType_Uint
	ParameterType_String
	ParameterType_Bool
	ParameterType_Choice
)

// A parameter that can be configured by the user
type Parameter struct {
	Name                 string
	ID                   string
	Description          string
	Type                 ParameterType
	Default              interface{}
	AffectsContainers    []ContainerID
	EnvironmentVariables []string
	CanBeBlank           bool
	OverwriteOnUpgrade   bool
}

// The value for a parameter
type Setting struct {
	Parameter    *Parameter
	Value        interface{}
	UsingDefault bool
}

// Configuration for the Execution client
type ExecutionConfig struct {
	ReconnectDelay *Parameter

	// External clients (Hybrid mode)
	UseExternalClient     *Parameter
	ExternalClientHttpUrl *Parameter
	ExternalClientWsUrl   *Parameter

	// Local clients (Docker mode)
	Client       *Parameter
	ClientConfig interface{}
}
