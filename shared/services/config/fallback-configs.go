package config

import (
	"github.com/rocket-pool/smartnode/shared/types/config"
)

// Configuration for fallback Lighthouse
type FallbackNormalConfig struct {
	Title string `yaml:"-"`

	// The URL of the Execution Client HTTP endpoint
	EcHttpUrl config.Parameter `yaml:"ecHttpUrl,omitempty"`

	// The URL of the Beacon Node HTTP endpoint
	CcHttpUrl config.Parameter `yaml:"ccHttpUrl,omitempty"`
}

// Configuration for fallback Prysm
type FallbackPrysmConfig struct {
	Title string `yaml:"-"`

	// The URL of the Execution Client HTTP endpoint
	EcHttpUrl config.Parameter `yaml:"ecHttpUrl,omitempty"`

	// The URL of the Beacon Node HTTP endpoint
	CcHttpUrl config.Parameter `yaml:"ccHttpUrl,omitempty"`

	// The URL of the JSON-RPC endpoint for the Validator client
	JsonRpcUrl config.Parameter `yaml:"jsonRpcUrl,omitempty"`
}

// Generates a new FallbackNormalConfig configuration
func NewFallbackNormalConfig(cfg *RocketPoolConfig) *FallbackNormalConfig {
	return &FallbackNormalConfig{
		Title: "Fallback Client Settings",

		EcHttpUrl: config.Parameter{
			ID:                 "ecHttpUrl",
			Name:               "Execution Client URL",
			Description:        "The URL of the HTTP API endpoint for your fallback Execution client.\n\nNOTE: If you are running it on the same machine as the Smartnode, addresses like `localhost` and `127.0.0.1` will not work due to Docker limitations. Enter your machine's LAN IP address instead.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Api, config.ContainerID_Node, config.ContainerID_Watchtower},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		CcHttpUrl: config.Parameter{
			ID:                 "ccHttpUrl",
			Name:               "Beacon Node URL",
			Description:        "The URL of the HTTP Beacon API endpoint for your fallback Consensus client.\n\nNOTE: If you are running it on the same machine as the Smartnode, addresses like `localhost` and `127.0.0.1` will not work due to Docker limitations. Enter your machine's LAN IP address instead.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Api, config.ContainerID_Node, config.ContainerID_Validator, config.ContainerID_Watchtower},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},
	}
}

// Generates a new FallbackPrysmConfig configuration
func NewFallbackPrysmConfig(cfg *RocketPoolConfig) *FallbackPrysmConfig {
	return &FallbackPrysmConfig{
		Title: "Fallback Prysm Settings",

		EcHttpUrl: config.Parameter{
			ID:                 "ecHttpUrl",
			Name:               "Execution Client URL",
			Description:        "The URL of the HTTP API endpoint for your fallback Execution client.\n\nNOTE: If you are running it on the same machine as the Smartnode, addresses like `localhost` and `127.0.0.1` will not work due to Docker limitations. Enter your machine's LAN IP address instead.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Api, config.ContainerID_Node, config.ContainerID_Watchtower},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		CcHttpUrl: config.Parameter{
			ID:                 "ccHttpUrl",
			Name:               "Beacon Node HTTP URL",
			Description:        "The URL of the HTTP Beacon API endpoint for your fallback Prysm client.\n\nNOTE: If you are running it on the same machine as the Smartnode, addresses like `localhost` and `127.0.0.1` will not work due to Docker limitations. Enter your machine's LAN IP address instead.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Api, config.ContainerID_Node, config.ContainerID_Validator, config.ContainerID_Watchtower},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		JsonRpcUrl: config.Parameter{
			ID:                 "jsonRpcUrl",
			Name:               "Beacon Node JSON-RPC URL",
			Description:        "The URL of the JSON-RPC API endpoint for your fallback client. Prysm's validator client will need this in order to connect to it.\n\nNOTE: If you are running it on the same machine as the Smartnode, addresses like `localhost` and `127.0.0.1` will not work due to Docker limitations. Enter your machine's LAN IP address instead.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth1},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},
	}
}

// Get the config.Parameters for this config
func (cfg *FallbackNormalConfig) GetParameters() []*config.Parameter {
	return []*config.Parameter{
		&cfg.EcHttpUrl,
		&cfg.CcHttpUrl,
	}
}

// Get the config.Parameters for this config
func (cfg *FallbackPrysmConfig) GetParameters() []*config.Parameter {
	return []*config.Parameter{
		&cfg.EcHttpUrl,
		&cfg.CcHttpUrl,
		&cfg.JsonRpcUrl,
	}
}

// The title for the config
func (config *FallbackNormalConfig) GetConfigTitle() string {
	return config.Title
}

// The title for the config
func (config *FallbackPrysmConfig) GetConfigTitle() string {
	return config.Title
}
