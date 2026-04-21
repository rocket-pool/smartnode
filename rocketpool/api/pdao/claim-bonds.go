package pdao

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/rocket-pool/smartnode/bindings/dao/protocol"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canClaimBonds(c *cli.Command, proposalId uint64, indices []uint64) (*api.PDAOCanClaimBondsResponse, error) {
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
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Response
	response := api.PDAOCanClaimBondsResponse{}

	// Sync
	var wg errgroup.Group
	var propState types.ProtocolDaoProposalState

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
		var err error
		propState, err = protocol.GetProposalState(rp, proposalId, nil)
		return err
	})

	// Get the proposer
	wg.Go(func() error {
		proposer, err := protocol.GetProposalProposer(rp, proposalId, nil)
		response.IsProposer = (proposer == nodeAccount.Address)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Check if it's in the right state
	if response.IsProposer {
		response.InvalidState = (propState < types.ProtocolDaoProposalState_QuorumNotMet)
	} else {
		response.InvalidState = (propState == types.ProtocolDaoProposalState_Pending)
	}

	// Verify
	response.CanClaim = !response.DoesNotExist && !response.InvalidState
	if !response.CanClaim {
		return &response, nil
	}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	var gasInfo rocketpool.GasInfo

	if response.IsProposer {
		gasInfo, err = protocol.EstimateClaimBondProposerGas(rp, proposalId, indices, opts)
	} else {
		gasInfo, err = protocol.EstimateClaimBondChallengerGas(rp, proposalId, indices, opts)
	}
	if err != nil {
		return nil, err
	}

	// Update & return response
	response.GasInfo = gasInfo
	return &response, nil
}

func claimBonds(c *cli.Command, isProposer bool, proposalId uint64, indices []uint64, opts *bind.TransactOpts) (*api.PDAOClaimBondsResponse, error) {
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
	response := api.PDAOClaimBondsResponse{}

	// Claim bonds
	if isProposer {
		response.TxHash, err = protocol.ClaimBondProposer(rp, proposalId, indices, opts)
	} else {
		response.TxHash, err = protocol.ClaimBondChallenger(rp, proposalId, indices, opts)
	}
	if err != nil {
		return nil, err
	}

	// Return response
	return &response, nil
}
