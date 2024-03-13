package node

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
)

// ===============
// === Factory ===
// ===============

type nodeClearSnapshotDelegateContextFactory struct {
	handler *NodeHandler
}

func (f *nodeClearSnapshotDelegateContextFactory) Create(args url.Values) (*nodeClearSnapshotDelegateContext, error) {
	c := &nodeClearSnapshotDelegateContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *nodeClearSnapshotDelegateContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*nodeClearSnapshotDelegateContext, types.TxInfoData](
		router, "snapshot-delegate/clear", f, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeClearSnapshotDelegateContext struct {
	handler *NodeHandler
}

func (c *nodeClearSnapshotDelegateContext) PrepareData(data *types.TxInfoData, opts *bind.TransactOpts) error {
	sp := c.handler.serviceProvider
	cfg := sp.GetConfig()
	snapshot := sp.GetSnapshotDelegation()
	idHash := cfg.GetVotingSnapshotID()

	var err error
	data.TxInfo, err = snapshot.ClearDelegate(idHash, opts)
	if err != nil {
		return fmt.Errorf("error getting TX info for ClearDelegate: %w", err)
	}
	return nil
}
