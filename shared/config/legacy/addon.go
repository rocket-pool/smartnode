package config

// Interface for Smartnode addons
type SmartnodeAddon interface {
	GetName() string
	GetDescription() string
	GetConfig() Config
	GetContainerName() string
	GetContainerTag() string
	GetEnabledParameter() *Parameter
}
