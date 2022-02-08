package config

type ContainerID int
type Network int
type Mode int
type ParameterType int

// Enum to describe which container(s) a parameter impacts, so the Smartnode knows which
// ones to restart upon a settings change
const (
	ContainerID_Unknown ContainerID = iota
	ContainerID_Api
	ContainerID_Node
	ContainerID_Watchtower
	ContainerID_Eth1
	ContainerID_Eth1Fallback
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

// Enum to describe the mode for a client - local (Docker Mode) or external (Hybrid Mode)
const (
	Mode_Unknown Mode = iota
	Mode_Local
	Mode_External
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
	ID                   string
	Name                 string
	Description          string
	Type                 ParameterType
	Default              interface{}
	Advanced             bool
	AffectsContainers    []ContainerID
	EnvironmentVariables []string
	CanBeBlank           bool
	OverwriteOnUpgrade   bool
	Options              []ParameterOption
	Value                interface{}
}

// A single option in a choice parameter
type ParameterOption struct {
	ID          string
	Name        string
	Description string
	Value       interface{}
}
