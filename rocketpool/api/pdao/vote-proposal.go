package pdao

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/dao"
	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/types"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
)

func canVoteOnProposal(c *cli.Context, proposalId uint64, voteDirection types.VoteDirection) (*api.CanVoteOnPDAOProposalResponse, error) {

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
	response := api.CanVoteOnPDAOProposalResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Data
	var wg errgroup.Group
	var proposalBlock uint32

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

	// Get the block used by the proposal
	wg.Go(func() error {
		var err error
		proposalBlock, err = protocol.GetProposalBlock(rp, proposalId, nil)
		return err
	})

	// Get gas estimate
	wg.Go(func() error {
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return err
		}
		gasInfo, err := protocol.EstimateVoteOnProposalGas(rp, proposalId, support, opts)
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
	votingPower, err := network.GetVotingPower(rp, nodeAccount.Address, proposalBlock, nil)
	if err != nil {
		return nil, err
	}
	response.VotingPower = votingPower
	response.InsufficientPower = (votingPower.Cmp(common.Big0) > 0)

	// Update & return response
	response.CanVote = !(response.DoesNotExist || response.InvalidState || response.InsufficientPower || response.AlreadyVoted)
	return &response, nil

}

func voteOnProposal(c *cli.Context, proposalId uint64, support bool) (*api.VoteOnPDAOProposalResponse, error) {

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
	response := api.VoteOnPDAOProposalResponse{}

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

	// Vote on proposal
	hash, err := protocol.VoteOnProposal(rp, proposalId, support, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}
