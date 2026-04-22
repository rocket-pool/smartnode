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

func canVoteOnProposal(c *cli.Command, proposalId uint64) (*api.SecurityCanVoteOnProposalResponse, error) {

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
	response := api.SecurityCanVoteOnProposalResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Data
	var wg errgroup.Group
	var memberJoinedTime uint64
	var proposalCreatedTime uint64

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
			response.InvalidState = (proposalState != rptypes.Active)
		}
		return err
	})

	// Check if member has already voted
	wg.Go(func() error {
		hasVoted, err := dao.GetProposalMemberVoted(rp, proposalId, nodeAccount.Address, nil)
		if err == nil {
			response.AlreadyVoted = hasVoted
		}
		return err
	})

	// Get member joined time
	wg.Go(func() error {
		var err error
		memberJoinedTime, err = security.GetMemberJoinedTime(rp, nodeAccount.Address, nil)
		return err
	})

	// Get proposal created time
	wg.Go(func() error {
		var err error
		proposalCreatedTime, err = dao.GetProposalCreatedTime(rp, proposalId, nil)
		return err
	})

	// Get gas estimate
	wg.Go(func() error {
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return err
		}
		gasInfo, err := security.EstimateVoteOnProposalGas(rp, proposalId, false, opts)
		if err == nil {
			response.GasInfo = gasInfo
		}
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Check data
	response.JoinedAfterCreated = (memberJoinedTime >= proposalCreatedTime)

	// Update & return response
	response.CanVote = !response.DoesNotExist && !response.InvalidState && !response.JoinedAfterCreated && !response.AlreadyVoted
	return &response, nil

}

func voteOnProposal(c *cli.Command, proposalId uint64, support bool, opts *bind.TransactOpts) (*api.SecurityVoteOnProposalResponse, error) {

	// Get services
	if err := services.RequireNodeSecurityMember(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.SecurityVoteOnProposalResponse{}

	// Vote on proposal
	hash, err := security.VoteOnProposal(rp, proposalId, support, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}
