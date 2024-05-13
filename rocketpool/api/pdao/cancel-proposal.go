package pdao

import (
	"bytes"
	"fmt"

	"github.com/rocket-pool/rocketpool-go/dao"
	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"
)

func canCancelProposal(c *cli.Context, proposalId uint64) (*api.CanCancelPDAOProposalResponse, error) {

	// Get services
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.CanCancelPDAOProposalResponse{}

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
			response.InvalidState = !(proposalState == rptypes.Pending || proposalState == rptypes.Active)
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
		gasInfo, err := protocol.EstimateCancelProposalGas(rp, proposalId, opts)
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
	response.CanCancel = !(response.DoesNotExist || response.InvalidState || response.InvalidProposer)
	return &response, nil

}

func cancelProposal(c *cli.Context, proposalId uint64) (*api.CancelPDAOProposalResponse, error) {

	// Get services
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.CancelPDAOProposalResponse{}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Cancel proposal
	hash, err := protocol.CancelProposal(rp, proposalId, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}
