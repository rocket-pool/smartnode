package odao

import (
	"bytes"

	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/bindings/dao"
	"github.com/rocket-pool/smartnode/bindings/dao/trustednode"
	rptypes "github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canCancelProposal(c *cli.Command, proposalId uint64) (*api.CanCancelTNDAOProposalResponse, error) {

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
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
	response := api.CanCancelTNDAOProposalResponse{}

	// Sync
	var wg errgroup.Group

	// Check proposal exists
	wg.Go(func() error {
		proposalCount, err := dao.GetProposalCount(rp, nil)
		if err == nil {
			response.DoesNotExist = (proposalId > proposalCount)
		}
		return err
	})

	// Check proposal state
	wg.Go(func() error {
		proposalState, err := dao.GetProposalState(rp, proposalId, nil)
		if err == nil {
			response.InvalidState = proposalState != rptypes.Pending && proposalState != rptypes.Active
		}
		return err
	})

	// Check proposer address
	wg.Go(func() error {
		nodeAccount, err := w.GetNodeAccount()
		if err != nil {
			return err
		}
		proposerAddress, err := dao.GetProposalProposerAddress(rp, proposalId, nil)
		if err == nil {
			response.InvalidProposer = !bytes.Equal(proposerAddress.Bytes(), nodeAccount.Address.Bytes())
		}
		return err
	})

	// Get gas estimate
	wg.Go(func() error {
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return err
		}
		gasLimits, err := trustednode.EstimateCancelProposalGas(rp, proposalId, opts)
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
	response.CanCancel = !response.DoesNotExist && !response.InvalidState && !response.InvalidProposer
	return &response, nil

}

func cancelProposal(c *cli.Command, proposalId uint64, t *snroute.TransactOpts) (*api.CancelTNDAOProposalResponse, error) {
	opts := t.Opts()

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.CancelTNDAOProposalResponse{}

	// Cancel proposal
	hash, err := trustednode.CancelProposal(rp, proposalId, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}

func canCancelProposalHandler(ctx snroute.Context) {
	id, err := parseUint64(ctx.Request, "id")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := canCancelProposal(ctx.Command(), id)
	response.WriteResponse(ctx.Writer, resp, err)
}

func cancelProposalHandler(ctx snroute.WriteContext) {
	id, err := parseUint64(ctx.Request, "id")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := cancelProposal(ctx.Command(), id, opts)
	response.WriteResponse(ctx.Writer, resp, err)
}
