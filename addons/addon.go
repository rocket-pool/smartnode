package addons

import (
	"github.com/rocket-pool/smartnode/shared/services/config"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

// Interface for Smartnode addons
type SmartnodeAddon interface {
	GetName() string
	GetDescription() string
	GetConfig(baseConfig *config.RocketPoolConfig) cfgtypes.Config
	GetContainerName() string
	GetContainerTag() string
	UpdateEnvVars(envVars map[string]string) error
}
