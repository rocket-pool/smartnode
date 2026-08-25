package security

import (
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/dao"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getProposals(c *cli.Command) (*api.SecurityProposalsResponse, error) {

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
	response := api.SecurityProposalsResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Get proposals
	proposals, err := dao.GetDAOProposalsWithMember(rp, "rocketDAOSecurityProposals", nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	response.Proposals = proposals

	// Return response
	return &response, nil

}

func getProposal(c *cli.Command, id uint64) (*api.SecurityProposalResponse, error) {

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
	response := api.SecurityProposalResponse{}

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

func proposalsHandler(ctx snroute.Context) {
	resp, err := getProposals(ctx.Command())
	response.WriteResponse(ctx.Writer, resp, err)
}

func proposalDetailsHandler(ctx snroute.Context) {
	id, err := parseUint64(ctx.Request, "id")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := getProposal(ctx.Command(), id)
	response.WriteResponse(ctx.Writer, resp, err)
}
