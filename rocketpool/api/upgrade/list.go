package upgrade

import (
	"github.com/rocket-pool/smartnode/bindings/dao/upgrades"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getUpgradeProposals(c *cli.Context) (*api.TNDAOGetUpgradeProposalsResponse, error) {

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
	response := api.TNDAOGetUpgradeProposalsResponse{}

	// Get upgradeProposals
	upgradeProposals, err := upgrades.GetUpgradeProposals(rp, nil)
	if err != nil {
		return nil, err
	}
	response.Proposals = upgradeProposals

	// Return response
	return &response, nil

}
