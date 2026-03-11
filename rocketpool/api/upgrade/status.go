package upgrade

import (
	"github.com/rocket-pool/smartnode/bindings/dao/upgrades"
	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getStatus(c *cli.Command) (*api.TNDAOUpgradeStatusResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.TNDAOUpgradeStatusResponse{}

	// Sync
	var wg errgroup.Group

	// Get upgrade proposal count
	upgradeProposalCount, err := upgrades.GetTotalUpgradeProposals(rp, nil)
	if err != nil {
		return nil, err
	}
	response.UpgradeProposalCount = upgradeProposalCount

	// // Get upgrade proposal state
	// wg.Go(func() error {
	// 	upgradeProposalStates, err := getUpgradeProposalStates(rp)
	// 	if err == nil {
	// 		response.UpgradeProposalState = upgradeProposalState.String()
	// 	}
	// 	return err
	// })

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Return response
	return &response, nil

}
