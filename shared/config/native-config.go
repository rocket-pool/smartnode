package config

import (
	"path/filepath"

	"github.com/rocket-pool/node-manager-core/config"
	"github.com/rocket-pool/smartnode/shared/config/ids"
)

// Configuration for Native mode
type NativeConfig struct {
	// The command for restarting the validator container in native mode
	ValidatorRestartCommand config.Parameter[string]

	// The command for stopping the validator container in native mode
	ValidatorStopCommand config.Parameter[string]
}

// Generates a new native mode configuration
func NewNativeConfig(cfg *SmartNodeConfig) *NativeConfig {
	return &NativeConfig{
		ValidatorRestartCommand: config.Parameter[string]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 ids.NativeValidatorRestartCommandID,
				Name:               "VC Restart Script",
				Description:        "The absolute path to a custom script that will be invoked when the Smart Node needs to restart your validator client to load the new key after a minipool is staked.",
				AffectsContainers:  []config.ContainerID{config.ContainerID_Daemon},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[config.Network]string{
				config.Network_All: getDefaultValidatorRestartCommand(cfg),
			},
		},

		ValidatorStopCommand: config.Parameter[string]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 ids.NativeValidatorStopCommandID,
				Name:               "Validator Stop Command",
				Description:        "The absolute path to a custom script that will be invoked when the Smart Node needs to stop your validator client in case of emergency.",
				AffectsContainers:  []config.ContainerID{config.ContainerID_Daemon},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[config.Network]string{
				config.Network_All: getDefaultValidatorStopCommand(cfg),
			},
		},
	}
}

// The title for the config
func (cfg *NativeConfig) GetTitle() string {
	return "Native"
}

// Get the parameters for this config
func (cfg *NativeConfig) GetParameters() []config.IParameter {
	return []config.IParameter{
		&cfg.ValidatorRestartCommand,
		&cfg.ValidatorStopCommand,
	}
}

// Get the sections underneath this one
func (cfg *NativeConfig) GetSubconfigs() map[string]config.IConfigSection {
	return map[string]config.IConfigSection{}
}

func getDefaultValidatorRestartCommand(cfg *SmartNodeConfig) string {
	return filepath.Join(cfg.RocketPoolDirectory, "restart-vc.sh")
}

func getDefaultValidatorStopCommand(cfg *SmartNodeConfig) string {
	return filepath.Join(cfg.RocketPoolDirectory, "stop-validator.sh")
}
