package graffiti_wall_writer

import (
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

// Configuration for the Graffiti Wall Writer
type GraffitiWallWriterConfig struct {
	Title string `yaml:"-"`

	Enabled cfgtypes.Parameter `yaml:"enabled,omitempty"`

	InputURL cfgtypes.Parameter `yaml:"inputUrl,omitempty"`

	UpdateWallTime cfgtypes.Parameter `yaml:"updateWallTime,omitempty"`

	UpdateInputTime cfgtypes.Parameter `yaml:"updateInputTime,omitempty"`

	UpdatePixelTime cfgtypes.Parameter `yaml:"updatePixelTime,omitempty"`
}

// Creates a new configuration instance
func NewConfig() *GraffitiWallWriterConfig {
	return &GraffitiWallWriterConfig{
		Title: "Graffiti Wall Writer Settings",

		Enabled: cfgtypes.Parameter{
			ID:                   "enabled",
			Name:                 "Enabled",
			Description:          "Enable the Graffiti Wall Writer",
			Type:                 cfgtypes.ParameterType_Bool,
			Default:              map[cfgtypes.Network]interface{}{cfgtypes.Network_All: false},
			AffectsContainers:    []cfgtypes.ContainerID{ContainerID_GraffitiWallWriter, cfgtypes.ContainerID_Validator},
			EnvironmentVariables: []string{"ADDON_GWW_ENABLED"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		InputURL: cfgtypes.Parameter{
			ID:                   "inputUrl",
			Name:                 "Input URL",
			Description:          "URL or path to file to source pixeldata from",
			Type:                 cfgtypes.ParameterType_String,
			Default:              map[cfgtypes.Network]interface{}{cfgtypes.Network_All: ""},
			AffectsContainers:    []cfgtypes.ContainerID{ContainerID_GraffitiWallWriter},
			EnvironmentVariables: []string{"ADDON_GWW_INPUT_URL"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},

		UpdateWallTime: cfgtypes.Parameter{
			ID:                   "updateWallTime",
			Name:                 "Wall Update Interval",
			Description:          "The time, in seconds, between updating the beaconcha.in graffiti wall canvas",
			Type:                 cfgtypes.ParameterType_Uint,
			Default:              map[cfgtypes.Network]interface{}{cfgtypes.Network_All: uint64(600)},
			AffectsContainers:    []cfgtypes.ContainerID{ContainerID_GraffitiWallWriter},
			EnvironmentVariables: []string{"ADDON_GWW_UPDATE_WALL_TIME"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},

		UpdateInputTime: cfgtypes.Parameter{
			ID:                   "updateInputTime",
			Name:                 "Input Update Interval",
			Description:          "The time, in seconds, between input updates - only if remote URL is used. File will be instantly reloaded when changed.",
			Type:                 cfgtypes.ParameterType_Uint,
			Default:              map[cfgtypes.Network]interface{}{cfgtypes.Network_All: uint64(600)},
			AffectsContainers:    []cfgtypes.ContainerID{ContainerID_GraffitiWallWriter},
			EnvironmentVariables: []string{"ADDON_GWW_UPDATE_INPUT_TIME"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},

		UpdatePixelTime: cfgtypes.Parameter{
			ID:                   "updatePixelTime",
			Name:                 "Pixel Update Interval",
			Description:          "The time, in seconds, between output updates.",
			Type:                 cfgtypes.ParameterType_Uint,
			Default:              map[cfgtypes.Network]interface{}{cfgtypes.Network_All: uint64(60)},
			AffectsContainers:    []cfgtypes.ContainerID{ContainerID_GraffitiWallWriter},
			EnvironmentVariables: []string{"ADDON_GWW_UPDATE_PIXEL_TIME"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},
	}
}

// Get the parameters for this config
func (cfg *GraffitiWallWriterConfig) GetParameters() []*cfgtypes.Parameter {
	return []*cfgtypes.Parameter{
		&cfg.Enabled,
		&cfg.InputURL,
		&cfg.UpdateWallTime,
		&cfg.UpdateInputTime,
		&cfg.UpdatePixelTime,
	}
}

// The the title for the config
func (cfg *GraffitiWallWriterConfig) GetConfigTitle() string {
	return cfg.Title
}
