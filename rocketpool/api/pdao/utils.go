package pdao

import (
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/proposals"
)

// Constructs a pollard for the latest finalized block and saves it to disk
func createPollard(rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, bc beacon.Client) (uint32, []types.VotingTreeNode, error) {
	// Create a proposal manager
	propMgr, err := proposals.NewProposalManager(nil, cfg, rp, bc)
	if err != nil {
		return 0, nil, err
	}

	// Create the pollard
	blockNumber, pollardPtrs, err := propMgr.CreatePollardForProposal()
	if err != nil {
		return 0, nil, err
	}

	// Make a slice of nodes from their pointers
	pollard := make([]types.VotingTreeNode, len(pollardPtrs))
	for i := range pollardPtrs {
		pollard[i] = *pollardPtrs[i]
	}
	return blockNumber, pollard, nil
}

// Loads (or regenerates) the pollard for a proposal from a block number
func getPollard(rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, bc beacon.Client, blockNumber uint32) ([]types.VotingTreeNode, error) {
	// Create a proposal manager
	propMgr, err := proposals.NewProposalManager(nil, cfg, rp, bc)
	if err != nil {
		return nil, err
	}

	// Get the pollard
	pollardPtrs, err := propMgr.GetPollardForProposal(blockNumber)
	if err != nil {
		return nil, err
	}

	// Make a slice of nodes from their pointers
	pollard := make([]types.VotingTreeNode, len(pollardPtrs))
	for i := range pollardPtrs {
		pollard[i] = *pollardPtrs[i]
	}
	return pollard, nil
}
