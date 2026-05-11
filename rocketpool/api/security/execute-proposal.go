package security

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/bindings/dao"
	"github.com/rocket-pool/smartnode/bindings/dao/security"
	rptypes "github.com/rocket-pool/smartnode/bindings/types"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canExecuteProposal(c *cli.Command, proposalId uint64) (*api.SecurityCanExecuteProposalResponse, error) {

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
	response := api.SecurityCanExecuteProposalResponse{}

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
			response.InvalidState = (proposalState != rptypes.Succeeded)
		}
		return err
	})

	// Get gas estimate
	wg.Go(func() error {
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return err
		}
		gasInfo, err := security.EstimateExecuteProposalGas(rp, proposalId, opts)
		if err == nil {
			response.GasInfo = gasInfo
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

func executeProposal(c *cli.Command, proposalId uint64, opts *bind.TransactOpts) (*api.SecurityExecuteProposalResponse, error) {

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
	response := api.SecurityExecuteProposalResponse{}

	// Cancel proposal
	hash, err := security.ExecuteProposal(rp, proposalId, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}
