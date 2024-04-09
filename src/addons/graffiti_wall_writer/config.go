package gww

import (
	"fmt"

	"github.com/rocket-pool/node-manager-core/config"
	nmc_ids "github.com/rocket-pool/node-manager-core/config/ids"
	"github.com/rocket-pool/smartnode/v2/addons/graffiti_wall_writer/ids"
)

// Constants
const (
	containerTag string = "rocketpool/graffiti-wall-addon:v1.0.1"
)

// Configuration for the Graffiti Wall Writer
type GraffitiWallWriterConfig struct {
	Enabled config.Parameter[bool]

	InputUrl config.Parameter[string]

	UpdateWallTime config.Parameter[uint64]

	UpdateInputTime config.Parameter[uint64]

	UpdatePixelTime config.Parameter[uint64]

	// The Docker Hub tag
	ContainerTag config.Parameter[string]

	// Custom command line flags
	AdditionalFlags config.Parameter[string]
}

// Creates a new configuration instance
func NewConfig() *GraffitiWallWriterConfig {
	return &GraffitiWallWriterConfig{
		Enabled: config.Parameter[bool]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 ids.GwwEnabledID,
				Name:               "Enabled",
				Description:        "Enable the Graffiti Wall Writer",
				AffectsContainers:  []config.ContainerID{ContainerID_GraffitiWallWriter, config.ContainerID_ValidatorClient},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[config.Network]bool{
				config.Network_All: false,
			},
		},

		InputUrl: config.Parameter[string]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 ids.GwwInputUrlID,
				Name:               "Input URL",
				Description:        "URL or filepath for the input JSON file that contains the graffiti image to write to the wall. By default, this is the Rocket Pool logo.\n\nSee https://gist.github.com/RomiRand/dfa1b5286af3e926deff0be2746db2df for info on making your own images.\n\nNOTE: for local files, you must manually put the file into the `addons/gww` folder of your `rocketpool` directory, and then enter the name of it as `/gww/<filename>` here.",
				AffectsContainers:  []config.ContainerID{ContainerID_GraffitiWallWriter},
				CanBeBlank:         true,
				OverwriteOnUpgrade: false,
			},
			Default: map[config.Network]string{
				config.Network_All: "https://cdn-rocketpool.s3.us-west-2.amazonaws.com/graffiti.json",
			},
		},

		UpdateWallTime: config.Parameter[uint64]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 ids.GwwUpdateWallTimeID,
				Name:               "Wall Update Interval",
				Description:        "The time, in seconds, between updating the beaconcha.in graffiti wall canvas",
				AffectsContainers:  []config.ContainerID{ContainerID_GraffitiWallWriter},
				CanBeBlank:         true,
				OverwriteOnUpgrade: false,
			},
			Default: map[config.Network]uint64{
				config.Network_All: 600,
			},
		},

		UpdateInputTime: config.Parameter[uint64]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 ids.GwwUpdateInputTimeID,
				Name:               "Input Update Interval",
				Description:        "The time, in seconds, between input updates - only if remote URL is used. File will be instantly reloaded when changed.",
				AffectsContainers:  []config.ContainerID{ContainerID_GraffitiWallWriter},
				CanBeBlank:         true,
				OverwriteOnUpgrade: false,
			},
			Default: map[config.Network]uint64{
				config.Network_All: 600,
			},
		},

		UpdatePixelTime: config.Parameter[uint64]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 ids.GwwUpdatePixelTimeID,
				Name:               "Pixel Update Interval",
				Description:        "The time, in seconds, between output updates.",
				AffectsContainers:  []config.ContainerID{ContainerID_GraffitiWallWriter},
				CanBeBlank:         true,
				OverwriteOnUpgrade: false,
			},
			Default: map[config.Network]uint64{
				config.Network_All: 60,
			},
		},

		ContainerTag: config.Parameter[string]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 nmc_ids.ContainerTagID,
				Name:               "Container Tag",
				Description:        "The tag name of the container you want to use on Docker Hub.",
				AffectsContainers:  []config.ContainerID{ContainerID_GraffitiWallWriter},
				CanBeBlank:         false,
				OverwriteOnUpgrade: true,
			},
			Default: map[config.Network]string{
				config.Network_All: containerTag,
			},
		},

		AdditionalFlags: config.Parameter[string]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 nmc_ids.AdditionalFlagsID,
				Name:               "Additional Flags",
				Description:        "Additional custom command line flags you want to pass to the addon, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
				AffectsContainers:  []config.ContainerID{ContainerID_GraffitiWallWriter},
				CanBeBlank:         true,
				OverwriteOnUpgrade: false,
			},
			Default: map[config.Network]string{
				config.Network_All: "",
			},
		},
	}
}

// Get the title for the config
func (cfg *GraffitiWallWriterConfig) GetTitle() string {
	return "Graffiti Wall Writer"
}

// Get the parameters for this config
func (cfg *GraffitiWallWriterConfig) GetParameters() []config.IParameter {
	return []config.IParameter{
		&cfg.Enabled,
		&cfg.InputUrl,
		&cfg.UpdateWallTime,
		&cfg.UpdateInputTime,
		&cfg.UpdatePixelTime,
		&cfg.ContainerTag,
		&cfg.AdditionalFlags,
	}
}

// Get the sections underneath this one
func (cfg *GraffitiWallWriterConfig) GetSubconfigs() map[string]config.IConfigSection {
	return map[string]config.IConfigSection{}
}

func (gww *GraffitiWallWriter) GetContainerName() string {
	return fmt.Sprint(ContainerID_GraffitiWallWriter)
}
