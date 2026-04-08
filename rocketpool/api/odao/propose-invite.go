package odao

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/dao/trustednode"
	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canProposeInvite(c *cli.Command, memberAddress common.Address, memberId, memberUrl string) (*api.CanProposeTNDAOInviteResponse, error) {

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
	response := api.CanProposeTNDAOInviteResponse{}

	// Sync
	var wg errgroup.Group

	// Check if proposal cooldown is active
	wg.Go(func() error {
		nodeAccount, err := w.GetNodeAccount()
		if err != nil {
			return err
		}
		proposalCooldownActive, err := getProposalCooldownActive(rp, nodeAccount.Address)
		if err == nil {
			response.ProposalCooldownActive = proposalCooldownActive
		}
		return err
	})

	// Check if member exists
	wg.Go(func() error {
		memberExists, err := trustednode.GetMemberExists(rp, memberAddress, nil)
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
		message := fmt.Sprintf("invite %s (%s)", memberId, memberUrl)
		gasInfo, err := trustednode.EstimateProposeInviteMemberGas(rp, message, memberAddress, memberId, memberUrl, opts)
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
	response.CanPropose = !(response.ProposalCooldownActive || response.MemberAlreadyExists)
	return &response, nil

}

func proposeInvite(c *cli.Command, memberAddress common.Address, memberId, memberUrl string, opts *bind.TransactOpts) (*api.ProposeTNDAOInviteResponse, error) {

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ProposeTNDAOInviteResponse{}

	// Submit proposal
	message := fmt.Sprintf("invite %s (%s)", memberId, memberUrl)
	proposalId, hash, err := trustednode.ProposeInviteMember(rp, message, memberAddress, memberId, memberUrl, opts)
	if err != nil {
		return nil, err
	}
	response.ProposalId = proposalId
	response.TxHash = hash

	// Return response
	return &response, nil

}
