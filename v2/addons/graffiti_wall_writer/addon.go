package gww

import (
	"github.com/rocket-pool/node-manager-core/config"
)

const (
	FolderName                     string             = "gww"
	ContainerID_GraffitiWallWriter config.ContainerID = "addon_gww"
	GraffitiFilename               string             = "graffiti.txt"
)

type GraffitiWallWriter struct {
	cfg *GraffitiWallWriterConfig `yaml:"config,omitempty"`
}

func NewGraffitiWallWriter(cfg *GraffitiWallWriterConfig) *GraffitiWallWriter {
	return &GraffitiWallWriter{
		cfg: cfg,
	}
}

func (gww *GraffitiWallWriter) GetName() string {
	return "Graffiti Wall Writer"
}

func (gww *GraffitiWallWriter) GetDescription() string {
	return "This addon adds support for drawing on the Beaconcha.in graffiti wall (https://beaconcha.in/graffitiwall) by replacing your validator's static graffiti message with a special message indicating a pixel to draw on the wall.\n\nMade with love by BenV and RamiRond!"
}
