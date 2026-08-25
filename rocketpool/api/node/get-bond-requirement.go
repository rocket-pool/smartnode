package node

import (
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
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

func getBondRequirementHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		numValidators, err := strconv.ParseUint(r.URL.Query().Get("numValidators"), 10, 64)
		if err != nil {
			response.WriteErrorResponse(w, fmt.Errorf("invalid numValidators: %w", err))
			return
		}
		resp, err := getBondRequirement(c, numValidators)
		response.WriteResponse(w, resp, err)
	}
}
