package node

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getBondRequirement(c *cli.Command, numValidators uint64) (*api.GetBondRequirementResponse, error) {

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

func getBondRequirementHandler(ctx snroute.Context) {
	numValidators, err := strconv.ParseUint(ctx.Request.URL.Query().Get("numValidators"), 10, 64)
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, fmt.Errorf("invalid numValidators: %ctx.Writer", err))
		return
	}
	resp, err := getBondRequirement(ctx.Command(), numValidators)
	response.WriteResponse(ctx.Writer, resp, err)
}
