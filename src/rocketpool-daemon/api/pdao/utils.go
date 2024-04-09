package pdao

import (
	"context"

	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/rocketpool-go/v2/types"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/proposals"
	"github.com/rocket-pool/smartnode/v2/shared/config"
)

// Constructs a pollard for the latest finalized block and saves it to disk
func createPollard(context context.Context, rp *rocketpool.RocketPool, cfg *config.SmartNodeConfig, bc beacon.IBeaconClient) (uint32, []types.VotingTreeNode, error) {
	// Create a proposal manager
	propMgr, err := proposals.NewProposalManager(context, nil, cfg, rp, bc)
	if err != nil {
		return 0, nil, err
	}

	// Create the pollard
	blockNumber, pollardPtrs, err := propMgr.CreatePollardForProposal(context)
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
func getPollard(context context.Context, rp *rocketpool.RocketPool, cfg *config.SmartNodeConfig, bc beacon.IBeaconClient, blockNumber uint32) ([]types.VotingTreeNode, error) {
	// Create a proposal manager
	propMgr, err := proposals.NewProposalManager(context, nil, cfg, rp, bc)
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
