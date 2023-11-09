package addons

import (
	"github.com/rocket-pool/smartnode/addons/graffiti_wall_writer"
	"github.com/rocket-pool/smartnode/addons/rescue_node"
	"github.com/rocket-pool/smartnode/shared/types/addons"
)

func NewGraffitiWallWriter() addons.SmartnodeAddon {
	return graffiti_wall_writer.NewGraffitiWallWriter()
}

func NewRescueNode() addons.SmartnodeAddon {
	return rescue_node.NewRescueNode()
}
