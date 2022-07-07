package graffiti_wall_writer

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/types/addons"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

const (
	ContainerID_GraffitiWallWriter cfgtypes.ContainerID = "gww"

	// Internal variables
	graffitiWallOutputFile string = "/validators/graffiti_wall.json"
)

type GraffitiWallWriter struct {
	cfg *GraffitiWallWriterConfig `yaml:"config,omitempty"`
}

func NewGraffitiWallWriter() addons.SmartnodeAddon {
	return &GraffitiWallWriter{
		cfg: NewConfig(),
	}
}

func (gww *GraffitiWallWriter) GetName() string {
	return "Graffiti Wall Writer"
}

func (gww *GraffitiWallWriter) GetDescription() string {
	return "This addon adds support for drawing on the Beaconcha.in graffiti wall (https://beaconcha.in/graffitiwall) using by replacing your validator's static graffiti message with a special message indicating a pixel to draw on the wall.\n\nMade with love by BenV and RamiRond!"
}

func (gww *GraffitiWallWriter) GetConfig() cfgtypes.Config {
	return gww.cfg
}

func (gww *GraffitiWallWriter) GetContainerName() string {
	return fmt.Sprint(ContainerID_GraffitiWallWriter)
}

func (gww *GraffitiWallWriter) GetEnabledParameter() *cfgtypes.Parameter {
	return &gww.cfg.Enabled
}

func (gww *GraffitiWallWriter) GetContainerTag() string {
	return "" // NYI
}

func (gww *GraffitiWallWriter) UpdateEnvVars(envVars map[string]string) error {
	if gww.cfg.Enabled.Value == true {
		cfgtypes.AddParametersToEnvVars(gww.cfg.GetParameters(), envVars)
	}
	return nil
}
