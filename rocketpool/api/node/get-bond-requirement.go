package node

import (
	"math/big"

	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/urfave/cli"
)

func getBondRequirement(c *cli.Context, numValidators uint64) (*api.GetBondRequirementResponse, error) {

	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	response := api.GetBondRequirementResponse{}

	bondRequirement, err := node.GetBondRequirement(rp, big.NewInt(int64(numValidators)), nil)
	if err != nil {
		return nil, err
	}
	response.BondRequirement = bondRequirement

	return &response, nil
}
