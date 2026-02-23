package megapool

import (
	"fmt"

	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getNewValidatorBondRequirement(c *cli.Context) (*api.GetNewValidatorBondRequirementResponse, error) {

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
	response := api.GetNewValidatorBondRequirementResponse{}

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

	// Check if the node's megapool is deployed
	deployed, err := megapool.GetMegapoolDeployed(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}
	if !deployed {
		return nil, fmt.Errorf("The node does not have a megapool deployed")
	}

	// Get the megapool contract
	mp, err := megapool.NewMegaPoolV1(rp, megapoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Get the bond amount required for the megapool's next validator
	newValidatorBondRequirement, err := mp.GetNewValidatorBondRequirement(nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting new validator bond requirement: %w", err)
	}

	response.NewValidatorBondRequirement = newValidatorBondRequirement
	return &response, nil
}
