// ================
// NOTE: the snapshot binding isn't built for multicall yet so this uses the old-school method of single getters
// ================

package network

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/rocketpool/common/voting"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type networkProposalContextFactory struct {
	handler *NetworkHandler
}

func (f *networkProposalContextFactory) Create(args url.Values) (*networkProposalContext, error) {
	c := &networkProposalContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *networkProposalContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*networkProposalContext, api.NetworkDaoProposalsData](
		router, "dao-proposals", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type networkProposalContext struct {
	handler *NetworkHandler
}

func (c *networkProposalContext) PrepareData(data *api.NetworkDaoProposalsData, opts *bind.TransactOpts) error {
	sp := c.handler.serviceProvider
	rp := sp.GetRocketPool()
	cfg := sp.GetConfig()
	nodeAddress, _ := sp.GetWallet().GetAddress()
	snapshot := sp.GetSnapshotDelegation()

	// Requirements
	err := errors.Join(
		sp.RequireNodeRegistered(),
		sp.RequireSnapshot(),
	)
	if err != nil {
		return err
	}
	data.AccountAddress = nodeAddress

	// Get delegate address
	idHash := cfg.Smartnode.GetVotingSnapshotID()
	err = rp.Query(func(mc *batch.MultiCaller) error {
		snapshot.Delegation(mc, &data.VotingDelegate, nodeAddress, idHash)
		return nil
	}, nil)
	if err != nil {
		return fmt.Errorf("error getting voting delegate info: %w", err)
	}

	// Get snapshot proposals
	snapshotResponse, err := voting.GetSnapshotProposals(cfg, data.AccountAddress, data.VotingDelegate, true)
	if err != nil {
		return fmt.Errorf("error getting snapshot proposals: %w", err)
	}

	data.ActiveSnapshotProposals = snapshotResponse
	return nil
}
