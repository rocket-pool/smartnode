package pdao

import (
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/rocket-pool/smartnode/bindings/dao/protocol"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canDefeatProposal(c *cli.Command, proposalId uint64, index uint64) (*api.PDAOCanDefeatProposalResponse, error) {
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
	response := api.PDAOCanDefeatProposalResponse{}
	var creationTime time.Time
	var challengeWindow time.Duration

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
			response.AlreadyDefeated = (proposalState == types.ProtocolDaoProposalState_Destroyed)
		}
		return err
	})

	// Get the proposal creation time
	wg.Go(func() error {
		var err error
		creationTime, err = protocol.GetProposalCreationTime(rp, proposalId, nil)
		return err
	})

	// Get the proposal challenge window
	wg.Go(func() error {
		var err error
		challengeWindow, err = protocol.GetChallengePeriod(rp, proposalId, nil)
		return err
	})

	// Get the challenge state
	wg.Go(func() error {
		state, err := protocol.GetChallengeState(rp, proposalId, index, nil)
		if err == nil {
			response.InvalidChallengeState = (state != types.ChallengeState_Challenged)
		}
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Validate
	defeatStart := creationTime.Add(challengeWindow)
	response.StillInChallengeWindow = (time.Until(defeatStart) > 0)
	response.CanDefeat = !(response.DoesNotExist || response.AlreadyDefeated || response.InvalidChallengeState || response.StillInChallengeWindow)
	if !response.CanDefeat {
		return &response, nil
	}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := protocol.EstimateDefeatProposalGas(rp, proposalId, index, opts)
	if err != nil {
		return nil, err
	}

	// Update & return response
	response.GasInfo = gasInfo
	return &response, nil
}

func defeatProposal(c *cli.Command, proposalId uint64, index uint64, opts *bind.TransactOpts) (*api.PDAODefeatProposalResponse, error) {
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
	response := api.PDAODefeatProposalResponse{}

	// Execute proposal
	hash, err := protocol.DefeatProposal(rp, proposalId, index, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil
}
