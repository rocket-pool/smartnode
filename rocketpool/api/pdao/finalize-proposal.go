package pdao

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/rocket-pool/smartnode/bindings/dao/protocol"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canFinalizeProposal(c *cli.Command, proposalId uint64) (*api.PDAOCanFinalizeProposalResponse, error) {
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
	response := api.PDAOCanFinalizeProposalResponse{}

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
			response.InvalidState = (proposalState != types.ProtocolDaoProposalState_Vetoed)
		}
		return err
	})

	// Check proposal state
	wg.Go(func() error {
		var err error
		response.AlreadyFinalized, err = protocol.GetProposalIsFinalized(rp, proposalId, nil)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Validate
	response.CanFinalize = !(response.DoesNotExist || response.InvalidState || response.AlreadyFinalized)
	if !response.CanFinalize {
		return &response, nil
	}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := protocol.EstimateFinalizeGas(rp, proposalId, opts)
	if err != nil {
		return nil, err
	}

	// Update & return response
	response.GasInfo = gasInfo
	return &response, nil
}

func finalizeProposal(c *cli.Command, proposalId uint64, opts *bind.TransactOpts) (*api.PDAOFinalizeProposalResponse, error) {
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
	response := api.PDAOFinalizeProposalResponse{}

	// Execute proposal
	hash, err := protocol.Finalize(rp, proposalId, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil
}
