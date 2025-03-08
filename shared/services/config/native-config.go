package config

import (
	"path/filepath"

	"github.com/rocket-pool/smartnode/shared/types/config"
)

// Configuration for Native mode
type NativeConfig struct {
	Title string `yaml:"-"`

	// The URL of the EC HTTP endpoint
	EcHttpUrl config.Parameter `yaml:"ecHttpUrl,omitempty"`

	// The selected CC
	ConsensusClient config.Parameter `yaml:"consensusClient,omitempty"`

	// The URL of the CC HTTP endpoint
	CcHttpUrl config.Parameter `yaml:"ccHttpUrl,omitempty"`

	// The command for restarting the validator container in native mode
	ValidatorRestartCommand config.Parameter `yaml:"validatorRestartCommand,omitempty"`

	// The command for stopping the validator container in native mode
	ValidatorStopCommand config.Parameter `yaml:"validatorStopCommand,omitempty"`
}

// Generates a new Smartnode configuration
func NewNativeConfig(cfg *RocketPoolConfig) *NativeConfig {

	return &NativeConfig{
		Title: "Native Settings",

		EcHttpUrl: config.Parameter{
			ID:                 "ecHttpUrl",
			Name:               "Execution Client URL",
			Description:        "The URL of the HTTP RPC endpoint for your Execution client (e.g. http://localhost:8545).",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:  []config.ContainerID{},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		ConsensusClient: config.Parameter{
			ID:                 "consensusClient",
			Name:               "Consensus Client",
			Description:        "Select which Consensus client you are using / will use.",
			Type:               config.ParameterType_Choice,
			Default:            map[config.Network]interface{}{config.Network_All: config.ConsensusClient_Nimbus},
			AffectsContainers:  []config.ContainerID{},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
			Options: []config.ParameterOption{{
				Name:        "Lighthouse",
				Description: "Lighthouse is a Consensus client with a heavy focus on speed and security. The team behind it, Sigma Prime, is an information security and software engineering firm who have funded Lighthouse along with the Ethereum Foundation, Consensys, and private individuals. Lighthouse is built in Rust and offered under an Apache 2.0 License.",
				Value:       config.ConsensusClient_Lighthouse,
			}, {
				Name:        "Lodestar",
				Description: "Lodestar is the fifth open-source Ethereum consensus client. It is written in Typescript maintained by ChainSafe Systems. Lodestar, their flagship product, is a production-capable Beacon Chain and Validator Client uniquely situated as the go-to for researchers and developers for rapid prototyping and browser usage.",
				Value:       config.ConsensusClient_Lodestar,
			}, {
				Name:        "Nimbus",
				Description: "Nimbus is a Consensus client implementation that strives to be as lightweight as possible in terms of resources used. This allows it to perform well on embedded systems, resource-restricted devices -- including Raspberry Pis and mobile devices -- and multi-purpose servers.",
				Value:       config.ConsensusClient_Nimbus,
			}, {
				Name:        "Prysm",
				Description: "Prysm is a Go implementation of Ethereum Consensus protocol with a focus on usability, security, and reliability. Prysm is developed by Prysmatic Labs, a company with the sole focus on the development of their client. Prysm is written in Go and released under a GPL-3.0 license.",
				Value:       config.ConsensusClient_Prysm,
			}, {
				Name:        "Teku",
				Description: "PegaSys Teku (formerly known as Artemis) is a Java-based Ethereum 2.0 client designed & built to meet institutional needs and security requirements. PegaSys is an arm of ConsenSys dedicated to building enterprise-ready clients and tools for interacting with the core Ethereum platform. Teku is Apache 2 licensed and written in Java, a language notable for its maturity & ubiquity.",
				Value:       config.ConsensusClient_Teku,
			}},
		},

		CcHttpUrl: config.Parameter{
			ID:                 "ccHttpUrl",
			Name:               "Consensus Client URL",
			Description:        "The URL of the HTTP Beacon API endpoint for your Consensus client (e.g. http://localhost:5052).",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:  []config.ContainerID{},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		ValidatorRestartCommand: config.Parameter{
			ID:                 "validatorRestartCommand",
			Name:               "VC Restart Script",
			Description:        "The absolute path to a custom script that will be invoked when Rocket Pool needs to restart your validator container to load the new key after a minipool is staked.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: getDefaultValidatorRestartCommand(cfg)},
			AffectsContainers:  []config.ContainerID{},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		ValidatorStopCommand: config.Parameter{
			ID:                 "validatorStopCommand",
			Name:               "Validator Stop Command",
			Description:        "The absolute path to a custom script that will be invoked when Rocket Pool needs to stop your validator container in case of emergency. **For Native mode only.**",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: getDefaultValidatorStopCommand(cfg)},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Node},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},
	}

}

// Get the parameters for this config
func (cfg *NativeConfig) GetParameters() []*config.Parameter {
	return []*config.Parameter{
		&cfg.EcHttpUrl,
		&cfg.ConsensusClient,
		&cfg.CcHttpUrl,
		&cfg.ValidatorRestartCommand,
		&cfg.ValidatorStopCommand,
	}
}

func getDefaultValidatorRestartCommand(config *RocketPoolConfig) string {
	return filepath.Join(config.RocketPoolDirectory, "restart-vc.sh")
}

func getDefaultValidatorStopCommand(config *RocketPoolConfig) string {
	return filepath.Join(config.RocketPoolDirectory, "stop-validator.sh")
}

// The title for the config
func (cfg *NativeConfig) GetConfigTitle() string {
	return cfg.Title
}
