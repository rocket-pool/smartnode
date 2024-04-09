package config

import (
	"github.com/rocket-pool/node-manager-core/config"
	gww "github.com/rocket-pool/smartnode/v2/addons/graffiti_wall_writer"
	rn "github.com/rocket-pool/smartnode/v2/addons/rescue_node"
	"github.com/rocket-pool/smartnode/v2/shared/config/ids"
)

type AddonsConfig struct {
	GraffitiWallWriter *gww.GraffitiWallWriterConfig
	RescueNode         *rn.RescueNodeConfig
}

// Generates a new addons config
func NewAddonsConfig() *AddonsConfig {
	return &AddonsConfig{
		GraffitiWallWriter: gww.NewConfig(),
		RescueNode:         rn.NewConfig(),
	}
}

// The title for the config
func (cfg *AddonsConfig) GetTitle() string {
	return "Addons"
}

// Get the parameters for this config
func (cfg *AddonsConfig) GetParameters() []config.IParameter {
	return []config.IParameter{}
}

// Get the sections underneath this one
func (cfg *AddonsConfig) GetSubconfigs() map[string]config.IConfigSection {
	return map[string]config.IConfigSection{
		ids.AddonsGwwID:        cfg.GraffitiWallWriter,
		ids.AddonsRescueNodeID: cfg.RescueNode,
	}
}
