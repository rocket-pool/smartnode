package odao

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/dao/trustednode"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
)

func canLeave(c *cli.Context) (*api.CanLeaveTNDAOResponse, error) {

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
	response := api.CanLeaveTNDAOResponse{}

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

	// Check if members can leave the oracle DAO
	wg.Go(func() error {
		membersCanLeave, err := getMembersCanLeave(rp)
		if err == nil {
			response.InsufficientMembers = !membersCanLeave
		}
		return err
	})

	// Get gas estimate
	wg.Go(func() error {
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return err
		}
		gasInfo, err := trustednode.EstimateLeaveGas(rp, opts.From, opts)
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
	response.CanLeave = !(response.ProposalExpired || response.InsufficientMembers)
	return &response, nil

}

func leave(c *cli.Context, bondRefundAddress common.Address) (*api.LeaveTNDAOResponse, error) {

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
	response := api.LeaveTNDAOResponse{}

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

	// Leave
	hash, err := trustednode.Leave(rp, bondRefundAddress, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}
