package odao

import (
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/dao"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getProposals(c *cli.Command) (*api.TNDAOProposalsResponse, error) {

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
	response := api.TNDAOProposalsResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Get proposals
	proposals, err := dao.GetDAOProposalsWithMember(rp, "rocketDAONodeTrustedProposals", nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	response.Proposals = proposals

	// Return response
	return &response, nil

}

func getProposal(c *cli.Command, id uint64) (*api.TNDAOProposalResponse, error) {

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
	response := api.TNDAOProposalResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Get proposals
	proposal, err := dao.GetProposalDetailsWithMember(rp, id, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	response.Proposal = proposal

	// Return response
	return &response, nil

}
