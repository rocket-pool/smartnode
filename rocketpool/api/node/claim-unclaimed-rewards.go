package node

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/node"
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
	response.GasInfo, err = node.EstimateClaimUnclaimedRewards(rp, nodeAddress, opts)
	if err != nil {
		return nil, err
	}
	response.CanClaim = true

	return &response, nil

}

func claimUnclaimedRewards(c *cli.Command, nodeAddress common.Address, opts *bind.TransactOpts) (*api.ClaimUnclaimedRewardsResponse, error) {

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
