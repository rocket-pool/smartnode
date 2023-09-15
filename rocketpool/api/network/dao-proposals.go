package network

import (
	"fmt"

	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/smartnode/rocketpool/api/node"
	"github.com/rocket-pool/smartnode/rocketpool/common/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

type networkProposalHandler struct {
}

func NewAuctionBidHandler(vars map[string]string) (*networkProposalHandler, error) {
	h := &networkProposalHandler{}
	return h, nil
}

func (h *networkProposalHandler) CreateBindings(ctx *callContext) error {
	return nil
}

func (h *networkProposalHandler) GetState(ctx *callContext, mc *batch.MultiCaller) {
}

// NOTE: the snapshot binding isn't built for multicall yet so this uses the old-school method of single getters
func (h *networkProposalHandler) PrepareData(ctx *callContext, data *api.NetworkDAOProposalsResponse) error {
	nodeAddress := ctx.nodeAddress
	cfg := ctx.cfg
	data.AccountAddress = nodeAddress

	sp := services.GetServiceProvider()
	s := sp.GetSnapshotDelegation()
	if s == nil {
		return fmt.Errorf("snapshot voting is not available on this network")
	}

	// Get snapshot proposals
	snapshotResponse, err := node.GetSnapshotProposals(cfg.Smartnode.GetSnapshotApiDomain(), cfg.Smartnode.GetSnapshotID(), "active")
	if err != nil {
		return fmt.Errorf("error getting snapshot proposals: %w", err)
	}

	// Get delegate address
	idHash := cfg.Smartnode.GetVotingSnapshotID()
	data.VotingDelegate, err = s.Delegation(nil, nodeAddress, idHash)
	if err != nil {
		return fmt.Errorf("error getting voting delegate info: %w", err)
	}

	// Get voted proposals
	votedProposals, err := node.GetSnapshotVotedProposals(cfg.Smartnode.GetSnapshotApiDomain(), cfg.Smartnode.GetSnapshotID(), nodeAddress, data.VotingDelegate)
	if err != nil {
		return fmt.Errorf("error getting proposal votes: %w", err)
	}
	data.ProposalVotes = votedProposals.Data.Votes
	data.ActiveSnapshotProposals = snapshotResponse.Data.Proposals
	return nil
}
