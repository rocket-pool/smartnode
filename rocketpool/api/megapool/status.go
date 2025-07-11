package megapool

import (
	"fmt"
	"math/big"

	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/urfave/cli"
)

func getStatus(c *cli.Context) (*api.MegapoolStatusResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	if err := services.RequireBeaconClientSynced(c); err != nil {
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
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.MegapoolStatusResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	details, err := services.GetNodeMegapoolDetails(rp, bc, nodeAccount.Address)
	if err != nil {
		return nil, err
	}
	response.Megapool = details

	// Get latest delegate address
	delegate, err := rp.GetContract("rocketMegapoolDelegate", nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting latest minipool delegate contract: %w", err)
	}
	response.LatestDelegate = *delegate.Address

	// Return response
	return &response, nil
}

func calculateRewards(c *cli.Context, amount *big.Int) (*api.MegapoolRewardSplitResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	if err := services.RequireBeaconClientSynced(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.MegapoolRewardSplitResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Calculate the rewards split for a given amount
	response, err = services.CalculateRewards(rp, amount, nodeAccount.Address)
	if err != nil {
		return nil, fmt.Errorf("Error getting rewards split for amount %s: %w", amount, err)
	}

	//Return response
	return &response, nil
}

func calculatePendingRewards(c *cli.Context) (*api.MegapoolRewardSplitResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	if err := services.RequireBeaconClientSynced(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.MegapoolRewardSplitResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Get the megapool address
	megapoolAddress, err := megapool.GetMegapoolExpectedAddress(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	// Load the megapool
	mp, err := megapool.NewMegaPoolV1(rp, megapoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Calculate the rewards split for a given amount
	pendingRewards, err := mp.CalculatePendingRewards(nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting pending rewards: %w", err)
	}
	response.RewardSplit = pendingRewards

	//Return response
	return &response, nil
}
