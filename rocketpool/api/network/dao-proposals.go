// ================
// NOTE: the snapshot binding isn't built for multicall yet so this uses the old-school method of single getters
// ================

package network

import (
	"fmt"

	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/smartnode/rocketpool/api/node"
	"github.com/rocket-pool/smartnode/rocketpool/common/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type networkProposalContextFactory struct {
	h *NetworkHandler
}

func (f *networkProposalContextFactory) Create(vars map[string]string) (*networkProposalContext, error) {
	c := &networkProposalContext{
		h: f.h,
	}
	return c, nil
}

func (f *networkProposalContextFactory) Run(c *networkProposalContext) (*api.ApiResponse[api.NetworkDaoProposalsData], error) {
	return runNetworkCall[api.NetworkDaoProposalsData](c)
}

// ===============
// === Context ===
// ===============

type networkProposalContext struct {
	h *NetworkHandler
	*commonContext
}

func (c *networkProposalContext) CreateBindings(ctx *commonContext) error {
	c.commonContext = ctx
	return nil
}

func (c *networkProposalContext) GetState(mc *batch.MultiCaller) {
}

func (c *networkProposalContext) PrepareData(data *api.NetworkDaoProposalsData) error {
	data.AccountAddress = c.nodeAddress

	sp := services.GetServiceProvider()
	s := sp.GetSnapshotDelegation()
	if s == nil {
		return fmt.Errorf("snapshot voting is not available on this network")
	}

	// Get snapshot proposals
	snapshotResponse, err := node.GetSnapshotProposals(c.cfg.Smartnode.GetSnapshotApiDomain(), c.cfg.Smartnode.GetSnapshotID(), "active")
	if err != nil {
		return fmt.Errorf("error getting snapshot proposals: %w", err)
	}

	// Get delegate address
	idHash := c.cfg.Smartnode.GetVotingSnapshotID()
	data.VotingDelegate, err = s.Delegation(nil, c.nodeAddress, idHash)
	if err != nil {
		return fmt.Errorf("error getting voting delegate info: %w", err)
	}

	// Get voted proposals
	votedProposals, err := node.GetSnapshotVotedProposals(c.cfg.Smartnode.GetSnapshotApiDomain(), c.cfg.Smartnode.GetSnapshotID(), c.nodeAddress, data.VotingDelegate)
	if err != nil {
		return fmt.Errorf("error getting proposal votes: %w", err)
	}
	data.ProposalVotes = votedProposals.Data.Votes
	data.ActiveSnapshotProposals = snapshotResponse.Data.Proposals
	return nil
}
