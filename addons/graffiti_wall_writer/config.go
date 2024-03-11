package gww

/*
import (
	"github.com/rocket-pool/smartnode/shared/types/config"
)

// Constants
const (
	containerTag string = "rocketpool/graffiti-wall-addon:v1.0.1"
)

// Configuration for the Graffiti Wall Writer
type GraffitiWallWriterConfig struct {
	Title string `yaml:"-"`

	Enabled config.Parameter `yaml:"enabled,omitempty"`

	InputURL config.Parameter `yaml:"inputUrl,omitempty"`

	UpdateWallTime config.Parameter `yaml:"updateWallTime,omitempty"`

	UpdateInputTime config.Parameter `yaml:"updateInputTime,omitempty"`

	UpdatePixelTime config.Parameter `yaml:"updatePixelTime,omitempty"`

	// The Docker Hub tag
	ContainerTag config.Parameter `yaml:"containerTag,omitempty"`

	// Custom command line flags
	AdditionalFlags config.Parameter `yaml:"additionalFlags,omitempty"`
}

// Creates a new configuration instance
func NewConfig() *GraffitiWallWriterConfig {
	return &GraffitiWallWriterConfig{
		Title: "Graffiti Wall Writer Settings",

		Enabled: config.Parameter{
			ID:                 "enabled",
			Name:               "Enabled",
			Description:        "Enable the Graffiti Wall Writer",
			Type:               config.ParameterType_Bool,
			Default:            map[config.Network]interface{}{config.Network_All: false},
			AffectsContainers:  []config.ContainerID{ContainerID_GraffitiWallWriter, config.ContainerID_Validator},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		InputURL: config.Parameter{
			ID:                 "inputUrl",
			Name:               "Input URL",
			Description:        "URL or filepath for the input JSON file that contains the graffiti image to write to the wall. By default, this is the Rocket Pool logo.\n\nSee https://gist.github.com/RomiRand/dfa1b5286af3e926deff0be2746db2df for info on making your own images.\n\nNOTE: for local files, you must manually put the file into the `addons/gww` folder of your `rocketpool` directory, and then enter the name of it as `/gww/<filename>` here.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: "https://cdn-rocketpool.s3.us-west-2.amazonaws.com/graffiti.json"},
			AffectsContainers:  []config.ContainerID{ContainerID_GraffitiWallWriter},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},

		UpdateWallTime: config.Parameter{
			ID:                 "updateWallTime",
			Name:               "Wall Update Interval",
			Description:        "The time, in seconds, between updating the beaconcha.in graffiti wall canvas",
			Type:               config.ParameterType_Uint,
			Default:            map[config.Network]interface{}{config.Network_All: uint64(600)},
			AffectsContainers:  []config.ContainerID{ContainerID_GraffitiWallWriter},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},

		UpdateInputTime: config.Parameter{
			ID:                 "updateInputTime",
			Name:               "Input Update Interval",
			Description:        "The time, in seconds, between input updates - only if remote URL is used. File will be instantly reloaded when changed.",
			Type:               config.ParameterType_Uint,
			Default:            map[config.Network]interface{}{config.Network_All: uint64(600)},
			AffectsContainers:  []config.ContainerID{ContainerID_GraffitiWallWriter},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},

		UpdatePixelTime: config.Parameter{
			ID:                 "updatePixelTime",
			Name:               "Pixel Update Interval",
			Description:        "The time, in seconds, between output updates.",
			Type:               config.ParameterType_Uint,
			Default:            map[config.Network]interface{}{config.Network_All: uint64(60)},
			AffectsContainers:  []config.ContainerID{ContainerID_GraffitiWallWriter},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},

		ContainerTag: config.Parameter{
			ID:                 "containerTag",
			Name:               "Container Tag",
			Description:        "The tag name of the container you want to use on Docker Hub.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: containerTag},
			AffectsContainers:  []config.ContainerID{ContainerID_GraffitiWallWriter},
			CanBeBlank:         false,
			OverwriteOnUpgrade: true,
		},

		AdditionalFlags: config.Parameter{
			ID:                 "additionalFlags",
			Name:               "Additional Flags",
			Description:        "Additional custom command line flags you want to pass to the addon, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:  []config.ContainerID{ContainerID_GraffitiWallWriter},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},
	}
}

// Get the parameters for this config
func (cfg *GraffitiWallWriterConfig) GetParameters() []*config.Parameter {
	return []*config.Parameter{
		&cfg.Enabled,
		&cfg.InputURL,
		&cfg.UpdateWallTime,
		&cfg.UpdateInputTime,
		&cfg.UpdatePixelTime,
		&cfg.ContainerTag,
		&cfg.AdditionalFlags,
	}
}

// The the title for the config
func (cfg *GraffitiWallWriterConfig) GetConfigTitle() string {
	return cfg.Title
}
*/
