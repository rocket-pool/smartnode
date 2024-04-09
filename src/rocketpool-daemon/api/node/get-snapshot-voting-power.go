package node

import (
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/rocketpool-go/v2/node"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/voting"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type nodeGetSnapshotVotingPowerContextFactory struct {
	handler *NodeHandler
}

func (f *nodeGetSnapshotVotingPowerContextFactory) Create(args url.Values) (*nodeGetSnapshotVotingPowerContext, error) {
	c := &nodeGetSnapshotVotingPowerContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *nodeGetSnapshotVotingPowerContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*nodeGetSnapshotVotingPowerContext, api.NodeGetSnapshotVotingPowerData](
		router, "get-snapshot-voting-power", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeGetSnapshotVotingPowerContext struct {
	handler *NodeHandler
	rp      *rocketpool.RocketPool

	node *node.Node
}

func (c *nodeGetSnapshotVotingPowerContext) PrepareData(data *api.NodeGetSnapshotVotingPowerData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	cfg := sp.GetConfig()
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

	data.VotingPower, err = voting.GetSnapshotVotingPower(cfg, nodeAddress)
	if err != nil {
		return types.ResponseStatus_Error, err
	}
	return types.ResponseStatus_Success, nil
}
