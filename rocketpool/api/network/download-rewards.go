package network

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/rewards"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/urfave/cli"
)

func downloadRewardsFile(c *cli.Context, interval uint64) (*api.DownloadRewardsFileResponse, error) {

	// Get services
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.DownloadRewardsFileResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Get the event info for the interval
	intervalInfo, err := rewards.GetIntervalInfo(rp, cfg, nodeAccount.Address, interval, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting interval %d info: %w", interval, err)
	}

	// Download the rewards file
	err = rewards.DownloadRewardsFile(cfg, interval, intervalInfo.CID, intervalInfo.MerkleRoot, true)
	if err != nil {
		return nil, err
	}

	// Return response
	return &response, nil
}
