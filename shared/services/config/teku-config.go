package config

import (
	"github.com/pbnjay/memory"
	"github.com/rocket-pool/smartnode/shared/types/config"
)

const (
	tekuTagTest         string = "consensys/teku:24.3.1"
	tekuTagProd         string = "consensys/teku:24.3.1"
	defaultTekuMaxPeers uint16 = 100
)

// Configuration for Teku
type TekuConfig struct {
	Title string `yaml:"-"`

	// Common parameters that Teku doesn't support and should be hidden
	UnsupportedCommonParams []string `yaml:"-"`

	// Max number of P2P peers to connect to
	JvmHeapSize config.Parameter `yaml:"jvmHeapSize,omitempty"`

	// The max number of P2P peers to connect to
	MaxPeers config.Parameter `yaml:"maxPeers,omitempty"`

	// The use slashing protection flag
	UseSlashingProtection config.Parameter `yaml:"useSlashingProtection,omitempty"`

	// The archive mode flag
	ArchiveMode config.Parameter `yaml:"archiveMode,omitempty"`

	// The Docker Hub tag for Lighthouse
	ContainerTag config.Parameter `yaml:"containerTag,omitempty"`

	// Custom command line flags for the BN
	AdditionalBnFlags config.Parameter `yaml:"additionalBnFlags,omitempty"`

	// Custom command line flags for the VC
	AdditionalVcFlags config.Parameter `yaml:"additionalVcFlags,omitempty"`
}

// Generates a new Teku configuration
func NewTekuConfig(cfg *RocketPoolConfig) *TekuConfig {
	return &TekuConfig{
		Title: "Teku Settings",

		UnsupportedCommonParams: []string{},

		JvmHeapSize: config.Parameter{
			ID:                 "jvmHeapSize",
			Name:               "JVM Heap Size",
			Description:        "The max amount of RAM, in MB, that Teku's JVM should limit itself to. Setting this lower will cause Teku to use less RAM, though it will always use more than this limit.\n\nUse 0 for automatic allocation.",
			Type:               config.ParameterType_Uint,
			Default:            map[config.Network]interface{}{config.Network_All: getTekuHeapSize()},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth1},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		MaxPeers: config.Parameter{
			ID:                 "maxPeers",
			Name:               "Max Peers",
			Description:        "The maximum number of peers your client should try to maintain. You can try lowering this if you have a low-resource system or a constrained network.",
			Type:               config.ParameterType_Uint16,
			Default:            map[config.Network]interface{}{config.Network_All: defaultTekuMaxPeers},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth2},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		ArchiveMode: config.Parameter{
			ID:                 "archiveMode",
			Name:               "Enable Archive Mode",
			Description:        "When enabled, Teku will run in \"archive\" mode which means it can recreate the state of the Beacon chain for a previous block. This is required for manually generating the Merkle rewards tree.\n\nIf you are sure you will never be manually generating a tree, you can disable archive mode.",
			Type:               config.ParameterType_Bool,
			Default:            map[config.Network]interface{}{config.Network_All: false},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth2},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		ContainerTag: config.Parameter{
			ID:          "containerTag",
			Name:        "Container Tag",
			Description: "The tag name of the Teku container you want to use on Docker Hub.",
			Type:        config.ParameterType_String,
			Default: map[config.Network]interface{}{
				config.Network_Mainnet: tekuTagProd,
				config.Network_Devnet:  tekuTagTest,
				config.Network_Holesky: tekuTagTest,
			},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth2, config.ContainerID_Validator},
			CanBeBlank:         false,
			OverwriteOnUpgrade: true,
		},

		UseSlashingProtection: config.Parameter{
			ID:                 "useSlashingProtection",
			Name:               "Use Validator Slashing Protection",
			Description:        "When enabled, Teku will use the Validator Slashing Protection feature. See https://docs.teku.consensys.io/how-to/prevent-slashing/detect-slashing for details.",
			Type:               config.ParameterType_Bool,
			Default:            map[config.Network]interface{}{config.Network_All: true},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth2, config.ContainerID_Validator},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		AdditionalBnFlags: config.Parameter{
			ID:                 "additionalBnFlags",
			Name:               "Additional Beacon Node Flags",
			Description:        "Additional custom command line flags you want to pass Teku's Beacon Node, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth2},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},

		AdditionalVcFlags: config.Parameter{
			ID:                 "additionalVcFlags",
			Name:               "Additional Validator Client Flags",
			Description:        "Additional custom command line flags you want to pass Teku's Validator Client, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Validator},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},
	}
}

// Get the parameters for this config
func (cfg *TekuConfig) GetParameters() []*config.Parameter {
	return []*config.Parameter{
		&cfg.JvmHeapSize,
		&cfg.MaxPeers,
		&cfg.ArchiveMode,
		&cfg.UseSlashingProtection,
		&cfg.ContainerTag,
		&cfg.AdditionalBnFlags,
		&cfg.AdditionalVcFlags,
	}
}

// Get the recommended heap size for Teku
func getTekuHeapSize() uint64 {
	totalMemoryGB := memory.TotalMemory() / 1024 / 1024 / 1024
	if totalMemoryGB < 9 {
		return 2048
	}
	return 0
}

// Get the common params that this client doesn't support
func (cfg *TekuConfig) GetUnsupportedCommonParams() []string {
	return cfg.UnsupportedCommonParams
}

// Get the Docker container name of the validator client
func (cfg *TekuConfig) GetValidatorImage() string {
	return cfg.ContainerTag.Value.(string)
}

// Get the Docker container name of the beacon client
func (cfg *TekuConfig) GetBeaconNodeImage() string {
	return cfg.ContainerTag.Value.(string)
}

// Get the name of the client
func (cfg *TekuConfig) GetName() string {
	return "Teku"
}

// The the title for the config
func (cfg *TekuConfig) GetConfigTitle() string {
	return cfg.Title
}
