package addons

import (
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

// Interface for Smartnode addons
type SmartnodeAddon interface {
	GetName() string
	GetDescription() string
	GetConfig() cfgtypes.Config
	GetContainerName() string
	GetContainerTag() string
	GetEnabledParameter() *cfgtypes.Parameter
}
