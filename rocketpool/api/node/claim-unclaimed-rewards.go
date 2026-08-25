package node

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canClaimUnclaimedRewards(c *cli.Command, nodeAddress common.Address) (*api.CanClaimUnclaimedRewardsResponse, error) {
	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.CanClaimUnclaimedRewardsResponse{}

	unclaimedRewards, err := node.GetUnclaimedRewardsRaw(rp, nodeAddress, nil)
	if err != nil {
		return nil, err
	}

	if unclaimedRewards != nil {
		response.CanClaim = false
	}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	response.GasLimits, err = node.EstimateClaimUnclaimedRewards(rp, nodeAddress, opts)
	if err != nil {
		return nil, err
	}
	response.CanClaim = true

	return &response, nil

}

func claimUnclaimedRewards(c *cli.Command, nodeAddress common.Address, t *snroute.TransactOpts) (*api.ClaimUnclaimedRewardsResponse, error) {
	opts := t.Opts()

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ClaimUnclaimedRewardsResponse{}

	// Claim unclaimed rewards
	hash, err := node.ClaimUnclaimedRewards(rp, nodeAddress, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}

func canClaimUnclaimedRewardsHandler(ctx snroute.Context) {
	nodeAddr := common.HexToAddress(ctx.Request.URL.Query().Get("nodeAddress"))
	resp, err := canClaimUnclaimedRewards(ctx.Command(), nodeAddr)
	response.WriteResponse(ctx.Writer, resp, err)
}

func claimUnclaimedRewardsHandler(ctx snroute.WriteContext) {
	nodeAddr := common.HexToAddress(ctx.Request.FormValue("nodeAddress"))
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := claimUnclaimedRewards(ctx.Command(), nodeAddr, opts)
	response.WriteResponse(ctx.Writer, resp, err)
}
