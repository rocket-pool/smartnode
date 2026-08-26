package security

import (
	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/bindings/dao/security"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canLeave(c *cli.Command) (*api.SecurityCanLeaveResponse, error) {

	// Get services
	if err := services.RequireNodeSecurityMember(c); err != nil {
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
	response := api.SecurityCanLeaveResponse{}

	// Sync
	var wg errgroup.Group

	// Check proposal actionable status
	wg.Go(func() error {
		nodeAccount, err := w.GetNodeAccount()
		if err != nil {
			return err
		}
		proposalActionable, err := getProposalIsActionable(rp, nodeAccount.Address, "leave")
		if err == nil {
			response.ProposalExpired = !proposalActionable
		}
		return err
	})

	// Get gas estimate
	wg.Go(func() error {
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return err
		}
		gasLimits, err := security.EstimateLeaveGas(rp, opts)
		if err == nil {
			response.GasLimits = gasLimits
		}
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Update & return response
	response.CanLeave = !(response.ProposalExpired)
	return &response, nil

}

func leave(c *cli.Command, t *snroute.TransactOpts) (*api.SecurityLeaveResponse, error) {
	opts := t.Opts()

	// Get services
	if err := services.RequireNodeSecurityMember(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.SecurityLeaveResponse{}

	// Leave
	hash, err := security.Leave(rp, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}

func canLeaveHandler(ctx snroute.Context) {
	resp, err := canLeave(ctx.Command())
	response.WriteResponse(ctx.Writer, resp, err)
}

func leaveHandler(ctx snroute.WriteContext) {
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := leave(ctx.Command(), opts)
	response.WriteResponse(ctx.Writer, resp, err)
}
