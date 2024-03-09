package config

import (
	"path/filepath"
)

// Configuration for Native mode
type NativeConfig struct {
	Title string `yaml:"-"`

	// The URL of the EC HTTP endpoint
	EcHttpUrl Parameter `yaml:"ecHttpUrl,omitempty"`

	// The selected CC
	ConsensusClient Parameter `yaml:"consensusClient,omitempty"`

	// The URL of the CC HTTP endpoint
	CcHttpUrl Parameter `yaml:"ccHttpUrl,omitempty"`

	// The command for restarting the validator container in native mode
	ValidatorRestartCommand Parameter `yaml:"validatorRestartCommand,omitempty"`

	// The command for stopping the validator container in native mode
	ValidatorStopCommand Parameter `yaml:"validatorStopCommand,omitempty"`
}

// Generates a new Smartnode configuration
func NewNativeConfig(cfg *RocketPoolConfig) *NativeConfig {

	return &NativeConfig{
		Title: "Native Settings",

		EcHttpUrl: Parameter{
			ID:                 "ecHttpUrl",
			Name:               "Execution Client URL",
			Description:        "The URL of the HTTP RPC endpoint for your Execution client (e.g. http://localhost:8545).",
			Type:               ParameterType_String,
			Default:            map[Network]interface{}{Network_All: ""},
			AffectsContainers:  []ContainerID{},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		ConsensusClient: Parameter{
			ID:                 "consensusClient",
			Name:               "Consensus Client",
			Description:        "Select which Consensus client you are using / will use.",
			Type:               ParameterType_Choice,
			Default:            map[Network]interface{}{Network_All: ConsensusClient_Nimbus},
			AffectsContainers:  []ContainerID{},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
			Options: []ParameterOption{{
				Name:        "Lighthouse",
				Description: "Lighthouse is a Consensus client with a heavy focus on speed and security. The team behind it, Sigma Prime, is an information security and software engineering firm who have funded Lighthouse along with the Ethereum Foundation, Consensys, and private individuals. Lighthouse is built in Rust and offered under an Apache 2.0 License.",
				Value:       ConsensusClient_Lighthouse,
			}, {
				Name:        "Lodestar",
				Description: "Lodestar is the fifth open-source Ethereum consensus client. It is written in Typescript maintained by ChainSafe Systems. Lodestar, their flagship product, is a production-capable Beacon Chain and Validator Client uniquely situated as the go-to for researchers and developers for rapid prototyping and browser usage.",
				Value:       ConsensusClient_Lodestar,
			}, {
				Name:        "Nimbus",
				Description: "Nimbus is a Consensus client implementation that strives to be as lightweight as possible in terms of resources used. This allows it to perform well on embedded systems, resource-restricted devices -- including Raspberry Pis and mobile devices -- and multi-purpose servers.",
				Value:       ConsensusClient_Nimbus,
			}, {
				Name:        "Prysm",
				Description: "Prysm is a Go implementation of Ethereum Consensus protocol with a focus on usability, security, and reliability. Prysm is developed by Prysmatic Labs, a company with the sole focus on the development of their client. Prysm is written in Go and released under a GPL-3.0 license.",
				Value:       ConsensusClient_Prysm,
			}, {
				Name:        "Teku",
				Description: "PegaSys Teku (formerly known as Artemis) is a Java-based Ethereum 2.0 client designed & built to meet institutional needs and security requirements. PegaSys is an arm of ConsenSys dedicated to building enterprise-ready clients and tools for interacting with the core Ethereum platform. Teku is Apache 2 licensed and written in Java, a language notable for its maturity & ubiquity.",
				Value:       ConsensusClient_Teku,
			}},
		},

		CcHttpUrl: Parameter{
			ID:                 "ccHttpUrl",
			Name:               "Consensus Client URL",
			Description:        "The URL of the HTTP Beacon API endpoint for your Consensus client (e.g. http://localhost:5052).",
			Type:               ParameterType_String,
			Default:            map[Network]interface{}{Network_All: ""},
			AffectsContainers:  []ContainerID{},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		ValidatorRestartCommand: Parameter{
			ID:                 "validatorRestartCommand",
			Name:               "VC Restart Script",
			Description:        "The absolute path to a custom script that will be invoked when Rocket Pool needs to restart your validator container to load the new key after a minipool is staked.",
			Type:               ParameterType_String,
			Default:            map[Network]interface{}{Network_All: getDefaultValidatorRestartCommand(cfg)},
			AffectsContainers:  []ContainerID{},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		ValidatorStopCommand: Parameter{
			ID:                 "validatorStopCommand",
			Name:               "Validator Stop Command",
			Description:        "The absolute path to a custom script that will be invoked when Rocket Pool needs to stop your validator container in case of emergency. **For Native mode only.**",
			Type:               ParameterType_String,
			Default:            map[Network]interface{}{Network_All: getDefaultValidatorStopCommand(cfg)},
			AffectsContainers:  []ContainerID{ContainerID_Node},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},
	}

}

// Get the parameters for this config
func (cfg *NativeConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&cfg.EcHttpUrl,
		&cfg.ConsensusClient,
		&cfg.CcHttpUrl,
		&cfg.ValidatorRestartCommand,
		&cfg.ValidatorStopCommand,
	}
}

func getDefaultValidatorRestartCommand(cfg *RocketPoolConfig) string {
	return filepath.Join(cfg.RocketPoolDirectory, "restart-vc.sh")
}

func getDefaultValidatorStopCommand(cfg *RocketPoolConfig) string {
	return filepath.Join(cfg.RocketPoolDirectory, "stop-validator.sh")
}

// The the title for the config
func (cfg *NativeConfig) GetConfigTitle() string {
	return cfg.Title
}
