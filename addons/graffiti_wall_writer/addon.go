package graffiti_wall_writer

import (
	"fmt"

	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

const (
	ContainerID_GraffitiWallWriter cfgtypes.ContainerID = "gww"

	// Internal variables
	graffitiWallOutputFile string = "/validators/graffiti_wall.json"
)

type GraffitiWallWriter struct {
	cfg *GraffitiWallWriterConfig
}

func (gww *GraffitiWallWriter) GetName() string {
	return "Graffiti Wall Writer"
}

func (gww *GraffitiWallWriter) GetDescription() string {
	return "This addon adds support for drawing on the Beaconcha.in graffiti wall (https://beaconcha.in/graffitiwall) using by replacing your validator's static graffiti message with a special message indicating a pixel to draw on the wall.\n\nMade with love by BenV and RamiRond!"
}

func (gww *GraffitiWallWriter) GetConfig() cfgtypes.Config {
	if gww.cfg == nil {
		gww.cfg = NewConfig()
	}
	return gww.cfg
}

func (gww *GraffitiWallWriter) GetContainerName() string {
	return fmt.Sprint(ContainerID_GraffitiWallWriter)
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
