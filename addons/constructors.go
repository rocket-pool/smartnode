package addons

import (
	"github.com/rocket-pool/smartnode/addons/apcupsd"
	"github.com/rocket-pool/smartnode/addons/graffiti_wall_writer"
	"github.com/rocket-pool/smartnode/shared/types/addons"
)

func NewGraffitiWallWriter() addons.SmartnodeAddon {
	return graffiti_wall_writer.NewGraffitiWallWriter()
}

func NewApcupsd() addons.SmartnodeAddon {
	return apcupsd.NewApcupsd()
}
