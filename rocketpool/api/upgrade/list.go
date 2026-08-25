package upgrade

import (
	"net/http"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/dao/upgrades"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/shared/services"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getUpgradeProposals(c *cli.Command) (*api.TNDAOGetUpgradeProposalsResponse, error) {

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

func getUpgradeProposalsHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := getUpgradeProposals(c)
		response.WriteResponse(w, resp, err)
	}
}
