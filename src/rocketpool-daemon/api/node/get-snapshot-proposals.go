package node

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/v2/node"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/voting"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type nodeGetSnapshotProposalsContextFactory struct {
	handler *NodeHandler
}

func (f *nodeGetSnapshotProposalsContextFactory) Create(args url.Values) (*nodeGetSnapshotProposalsContext, error) {
	c := &nodeGetSnapshotProposalsContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("active-only", args, input.ValidateBool, &c.activeOnly),
	}
	return c, errors.Join(inputErrs...)
}

func (f *nodeGetSnapshotProposalsContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*nodeGetSnapshotProposalsContext, api.NodeGetSnapshotProposalsData](
		router, "get-snapshot-proposals", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeGetSnapshotProposalsContext struct {
	handler *NodeHandler
	rp      *rocketpool.RocketPool

	activeOnly bool
	node       *node.Node
}

func (c *nodeGetSnapshotProposalsContext) PrepareData(data *api.NodeGetSnapshotProposalsData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	cfg := sp.GetConfig()
	rp := sp.GetRocketPool()
	snapshot := sp.GetSnapshotDelegation()
	nodeAddress, _ := sp.GetWallet().GetAddress()

	// Requirements
	status, err := sp.RequireNodeRegistered(c.handler.ctx)
	if err != nil {
		return status, err
	}
	err = sp.RequireSnapshot()
	if err != nil {
		return types.ResponseStatus_InvalidChainState, err
	}

	var delegate common.Address
	err = rp.Query(func(mc *batch.MultiCaller) error {
		snapshot.Delegation(mc, &delegate, nodeAddress, cfg.GetVotingSnapshotID())
		return nil
	}, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting snapshot delegate: %w", err)
	}

	data.Proposals, err = voting.GetSnapshotProposals(cfg, nodeAddress, delegate, c.activeOnly)
	return types.ResponseStatus_Success, err
}
