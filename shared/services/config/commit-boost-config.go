package config

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/types/config"
)

// Constants
const (
	CommitBoostConfigFile string = "cb_config.toml"
	commitBoostProdTag    string = "ghcr.io/commit-boost/pbs:v0.9.2"
	commitBoostTestTag    string = "ghcr.io/commit-boost/pbs:v0.9.2"
)

// Configuration for Commit-Boost's service
type CommitBoostConfig struct {
	// Ownership mode
	Mode config.Parameter `yaml:"mode,omitempty"`

	// The URL of an external MEV-Boost client
	ExternalUrl config.Parameter `yaml:"externalUrl"`

	// The Docker Hub tag for Commit-Boost
	ContainerTag config.Parameter `yaml:"containerTag,omitempty"`

	// Custom command line flags
	AdditionalFlags config.Parameter `yaml:"additionalFlags,omitempty"`

	parentConfig *RocketPoolConfig `yaml:"-"`

	// The port that Commit-Boost should serve its API on
	Port config.Parameter `yaml:"port,omitempty"`

	// Toggle for forwarding the HTTP port outside of Docker
	OpenRpcPort config.Parameter `yaml:"openRpcPort,omitempty"`
}

// Generates a new Commit-Boost PBS service configuration
func NewCommitBoostConfig(cfg *RocketPoolConfig) *CommitBoostConfig {
	portModes := config.PortModes("")

	return &CommitBoostConfig{
		parentConfig: cfg,

		Mode: config.Parameter{
			ID:                 "mode",
			Name:               "Commit-Boost Mode",
			Description:        "Choose whether to let the Smartnode manage your Commit-Boost instance (Locally Managed), or if you manage your own outside of the Smartnode stack (Externally Managed).",
			Type:               config.ParameterType_Choice,
			Default:            map[config.Network]interface{}{config.Network_All: config.Mode_Local},
			AffectsContainers:  []config.ContainerID{config.ContainerID_CommitBoost},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
			Options: []config.ParameterOption{{
				Name:        "Locally Managed",
				Description: "Allow the Smartnode to manage the Commit-Boost client for you",
				Value:       config.Mode_Local,
			}, {
				Name:        "Externally Managed",
				Description: "Use an existing Commit-Boost client that you manage on your own",
				Value:       config.Mode_External,
			}},
		},
		Port: config.Parameter{
			ID:                 "port",
			Name:               "Port",
			Description:        "The port that Commit-Boost should serve its API on.",
			Type:               config.ParameterType_Uint16,
			Default:            map[config.Network]interface{}{config.Network_All: uint16(18550)},
			AffectsContainers:  []config.ContainerID{config.ContainerID_CommitBoost},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		OpenRpcPort: config.Parameter{
			ID:                 "openRpcPort",
			Name:               "Expose API Port",
			Description:        "Expose the API port to other processes on your machine, or to your local network so other local machines can access Commit-Boost's API.",
			Type:               config.ParameterType_Choice,
			Default:            map[config.Network]interface{}{config.Network_All: config.RPC_Closed},
			AffectsContainers:  []config.ContainerID{config.ContainerID_CommitBoost},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
			Options:            portModes,
		},

		ContainerTag: config.Parameter{
			ID:                 "containerTag",
			Name:               "Container Tag",
			Description:        "The tag name of the Commit-Boost container you want to use.",
			AffectsContainers:  []config.ContainerID{config.ContainerID_CommitBoost},
			CanBeBlank:         false,
			OverwriteOnUpgrade: true,
			Type:               config.ParameterType_String,
			Default: map[config.Network]interface{}{
				config.Network_Mainnet: commitBoostProdTag,
				config.Network_Devnet:  commitBoostTestTag,
				config.Network_Testnet: commitBoostTestTag,
			},
		},
		AdditionalFlags: config.Parameter{
			ID:                 "additionalFlags",
			Name:               "Additional Flags",
			Description:        "Additional custom command line flags you want to pass to Commit-Boost, to take advantage of other settings that Hyperdrive's configuration doesn't cover.",
			AffectsContainers:  []config.ContainerID{config.ContainerID_CommitBoost},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
			Type:               config.ParameterType_String,
			Default: map[config.Network]interface{}{
				config.Network_All: "",
			},
		},
		ExternalUrl: config.Parameter{
			ID:          "externalUrl",
			Name:        "External URL",
			Description: "The URL of the external Commit-Boost client or provider",
			Type:        config.ParameterType_String,
			Default:     map[config.Network]interface{}{config.Network_All: ""},
		},
	}
}

// The title for the config
func (cfg *CommitBoostConfig) GetConfigTitle() string {
	return "Commit-Boost Settings"
}

// Get the Parameters for this config
func (cfg *CommitBoostConfig) GetParameters() []*config.Parameter {
	return []*config.Parameter{
		&cfg.ContainerTag,
		&cfg.AdditionalFlags,
	}
}

// Get the filename for the Commit-Boost PBS config
func (cfg *CommitBoostConfig) GetCommitBoostConfigFilename() string {
	return CommitBoostConfigFile
}

func (cfg *CommitBoostConfig) GetCommitBoostOpenPorts() string {
	portMode := cfg.OpenRpcPort.Value.(config.RPCMode)
	if !portMode.Open() {
		return ""
	}
	port := cfg.Port.Value.(uint16)
	return fmt.Sprintf("\"%s\"", portMode.DockerPortMapping(port))
}

// Get the chain name for the Commit-Boost config file
func (cfg *CommitBoostConfig) GetChainName(network config.Network) (string, error) {
	switch network {
	case config.Network_Mainnet:
		return "Mainnet", nil
	case config.Network_Devnet:
		return "Devnet", nil
	case config.Network_Testnet:
		return "Testnet", nil
	default:
		return "", fmt.Errorf("unsupported network %s for Commit-Boost PBS config", network)
	}
}
