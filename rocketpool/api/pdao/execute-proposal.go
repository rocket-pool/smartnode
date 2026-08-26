package pdao

import (
	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/bindings/dao/protocol"
	rptypes "github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canExecuteProposal(c *cli.Command, proposalId uint64) (*api.CanExecutePDAOProposalResponse, error) {

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
	response := api.CanExecutePDAOProposalResponse{}

	// Sync
	var wg errgroup.Group

	// Check proposal exists
	wg.Go(func() error {
		proposalCount, err := protocol.GetTotalProposalCount(rp, nil)
		if err == nil {
			response.DoesNotExist = (proposalId > proposalCount)
		}
		return err
	})

	// Check proposal state
	wg.Go(func() error {
		proposalState, err := protocol.GetProposalState(rp, proposalId, nil)
		if err == nil {
			response.InvalidState = (proposalState != rptypes.ProtocolDaoProposalState_Succeeded)
		}
		return err
	})

	// Get gas estimate
	wg.Go(func() error {
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return err
		}
		gasLimits, err := protocol.EstimateExecuteProposalGas(rp, proposalId, opts)
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
	response.CanExecute = !response.DoesNotExist && !response.InvalidState
	return &response, nil

}

func executeProposal(c *cli.Command, proposalId uint64, t *snroute.TransactOpts) (*api.ExecutePDAOProposalResponse, error) {
	opts := t.Opts()

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
	response := api.ExecutePDAOProposalResponse{}

	// Execute proposal
	hash, err := protocol.ExecuteProposal(rp, proposalId, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}

func canExecuteProposalHandler(ctx snroute.Context) {
	id, err := parseUint64Param(ctx.Request, "id")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := canExecuteProposal(ctx.Command(), id)
	response.WriteResponse(ctx.Writer, resp, err)
}

func executeProposalHandler(ctx snroute.WriteContext) {
	id, err := parseUint64Param(ctx.Request, "id")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := executeProposal(ctx.Command(), id, opts)
	response.WriteResponse(ctx.Writer, resp, err)
}
