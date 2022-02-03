package node

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
)

func canNodeClaimRpl(c *cli.Context) (*api.CanNodeClaimRplResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
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
	response := api.CanNodeClaimRplResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Check for rewards
	rewardsAmountWei, err := rewards.GetNodeClaimRewardsAmount(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting RPL rewards amount: %w", err)
	}
	response.RplAmount = rewardsAmountWei

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := rewards.EstimateClaimNodeRewardsGas(rp, opts)
	if err != nil {
		return nil, fmt.Errorf("Could not estimate the gas required to claim RPL: %w", err)
	}
	response.GasInfo = gasInfo

	return &response, nil
}

func nodeClaimRpl(c *cli.Context) (*api.NodeClaimRplResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
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
	response := api.NodeClaimRplResponse{}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Claim rewards
	hash, err := rewards.ClaimNodeRewards(rp, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}
