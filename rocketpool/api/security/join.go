package security

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/bindings/dao/security"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canJoin(c *cli.Command) (*api.SecurityCanJoinResponse, error) {

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
	response := api.SecurityCanJoinResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Data
	var wg errgroup.Group

	// Check proposal actionable status
	wg.Go(func() error {
		proposalActionable, err := getProposalIsActionable(rp, nodeAccount.Address, "invited")
		if err == nil {
			response.ProposalExpired = !proposalActionable
		}
		return err
	})

	// Check if already a member
	wg.Go(func() error {
		isMember, err := security.GetMemberExists(rp, nodeAccount.Address, nil)
		if err == nil {
			response.AlreadyMember = isMember
		}
		return err
	})
	// Get gas estimate
	wg.Go(func() error {
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return err
		}
		gasInfo, err := security.EstimateJoinGas(rp, opts)
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
	response.CanJoin = !response.ProposalExpired && !response.AlreadyMember
	return &response, nil

}

func join(c *cli.Command, opts *bind.TransactOpts) (*api.SecurityJoinResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.SecurityJoinResponse{}

	// Join
	hash, err := security.Join(rp, opts)
	if err != nil {
		return nil, err
	}

	response.TxHash = hash

	// Return response
	return &response, nil

}
