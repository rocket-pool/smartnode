package security

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/dao/security"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
)

func canProposeInvite(c *cli.Context, memberId string, memberAddress common.Address) (*api.SecurityCanProposeInviteResponse, error) {

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
	response := api.SecurityCanProposeInviteResponse{}

	// Sync
	var wg errgroup.Group

	// Check if member exists
	wg.Go(func() error {
		memberExists, err := security.GetMemberExists(rp, memberAddress, nil)
		if err == nil {
			response.MemberAlreadyExists = memberExists
		}
		return err
	})

	// Get gas estimate
	wg.Go(func() error {
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return err
		}
		message := fmt.Sprintf("invite %s (%s)", memberId, memberAddress)
		gasInfo, err := security.EstimateProposeInviteMemberGas(rp, message, memberId, memberAddress, opts)
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
	response.CanPropose = !(response.MemberAlreadyExists)
	return &response, nil

}

func proposeInvite(c *cli.Context, memberId string, memberAddress common.Address) (*api.SecurityProposeInviteResponse, error) {

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
	response := api.SecurityProposeInviteResponse{}

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

	// Submit proposal
	message := fmt.Sprintf("invite %s (%s)", memberId, memberAddress)
	proposalId, hash, err := security.ProposeInviteMember(rp, message, memberId, memberAddress, opts)
	if err != nil {
		return nil, err
	}
	response.ProposalId = proposalId
	response.TxHash = hash

	// Return response
	return &response, nil

}
