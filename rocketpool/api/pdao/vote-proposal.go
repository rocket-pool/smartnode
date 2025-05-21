package pdao

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/dao/protocol"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/proposals"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
)

func canVoteOnProposal(c *cli.Context, proposalId uint64, voteDirection types.VoteDirection) (*api.CanVoteOnPDAOProposalResponse, error) {
	// Get services
	cfg, err := services.GetConfig(c)
	if err != nil {
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
	bc, err := services.GetBeaconClient(c)
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
			response.InvalidState = (proposalState != types.ProtocolDaoProposalState_ActivePhase1)
		}
		return err
	})

	// Check if member has already voted
	wg.Go(func() error {
		voteDirection, err := protocol.GetAddressVoteDirection(rp, proposalId, nodeAccount.Address, nil)
		if err == nil {
			response.AlreadyVoted = (voteDirection != types.VoteDirection_NoVote)
		}
		return err
	})

	// Get the block used by the proposal
	wg.Go(func() error {
		var err error
		proposalBlock, err = protocol.GetProposalBlock(rp, proposalId, nil)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Get the proposal artifacts
	propMgr, err := proposals.NewProposalManager(nil, cfg, rp, bc)
	if err != nil {
		return nil, err
	}

	totalDelegatedVP, nodeIndex, proof, err := propMgr.GetArtifactsForVoting(proposalBlock, nodeAccount.Address)
	if err != nil {
		return nil, err
	}
	response.VotingPower = totalDelegatedVP

	// Check data
	response.InsufficientPower = (response.VotingPower.Cmp(common.Big0) == 0)
	response.CanVote = !(response.DoesNotExist || response.InvalidState || response.InsufficientPower || response.AlreadyVoted)
	if !response.CanVote {
		return &response, nil
	}

	// Simulate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := protocol.EstimateVoteOnProposalGas(rp, proposalId, voteDirection, totalDelegatedVP, nodeIndex, proof, opts)
	if err != nil {
		return nil, err
	}
	response.GasInfo = gasInfo

	// Update & return response
	return &response, nil
}

func voteOnProposal(c *cli.Context, proposalId uint64, voteDirection types.VoteDirection) (*api.VoteOnPDAOProposalResponse, error) {
	// Get services
	cfg, err := services.GetConfig(c)
	if err != nil {
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
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.VoteOnPDAOProposalResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Get the block used by the proposal
	proposalBlock, err := protocol.GetProposalBlock(rp, proposalId, nil)
	if err != nil {
		return nil, err
	}

	// Get the proposal artifacts
	propMgr, err := proposals.NewProposalManager(nil, cfg, rp, bc)
	if err != nil {
		return nil, err
	}
	totalDelegatedVP, nodeIndex, proof, err := propMgr.GetArtifactsForVoting(proposalBlock, nodeAccount.Address)
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Vote on proposal
	hash, err := protocol.VoteOnProposal(rp, proposalId, voteDirection, totalDelegatedVP, nodeIndex, proof, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil
}
