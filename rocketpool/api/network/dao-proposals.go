// ================
// NOTE: the snapshot binding isn't built for multicall yet so this uses the old-school method of single getters
// ================

package network

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/rocketpool/api/node"
	"github.com/rocket-pool/smartnode/rocketpool/common/contracts"
	"github.com/rocket-pool/smartnode/shared/config"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type networkProposalContextFactory struct {
	handler *NetworkHandler
}

func (f *networkProposalContextFactory) Create(vars map[string]string) (*networkProposalContext, error) {
	c := &networkProposalContext{
		handler: f.handler,
	}
	return c, nil
}

// ===============
// === Context ===
// ===============

type networkProposalContext struct {
	handler     *NetworkHandler
	rp          *rocketpool.RocketPool
	cfg         *config.RocketPoolConfig
	nodeAddress common.Address
	snapshot    *contracts.SnapshotDelegation
}

func (c *networkProposalContext) PrepareData(data *api.NetworkDaoProposalsData, opts *bind.TransactOpts) error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.cfg = sp.GetConfig()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()
	c.snapshot = sp.GetSnapshotDelegation()

	// Requirements
	err := errors.Join(
		sp.RequireNodeRegistered(),
		sp.RequireSnapshot(),
	)
	if err != nil {
		return err
	}

	data.AccountAddress = c.nodeAddress

	// Get snapshot proposals
	snapshotResponse, err := node.GetSnapshotProposals(c.cfg.Smartnode.GetSnapshotApiDomain(), c.cfg.Smartnode.GetSnapshotID(), "active")
	if err != nil {
		return fmt.Errorf("error getting snapshot proposals: %w", err)
	}

	// Get delegate address
	idHash := c.cfg.Smartnode.GetVotingSnapshotID()
	data.VotingDelegate, err = c.snapshot.Delegation(nil, c.nodeAddress, idHash)
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
