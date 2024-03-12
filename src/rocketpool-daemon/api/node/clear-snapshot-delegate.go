package node

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"

	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
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
	server.RegisterQuerylessGet[*nodeClearSnapshotDelegateContext, api.TxInfoData](
		router, "snapshot-delegate/clear", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeClearSnapshotDelegateContext struct {
	handler *NodeHandler
}

func (c *nodeClearSnapshotDelegateContext) PrepareData(data *api.TxInfoData, opts *bind.TransactOpts) error {
	sp := c.handler.serviceProvider
	cfg := sp.GetConfig()
	snapshot := sp.GetSnapshotDelegation()
	idHash := cfg.Smartnode.GetVotingSnapshotID()

	var err error
	data.TxInfo, err = snapshot.ClearDelegate(idHash, opts)
	if err != nil {
		return fmt.Errorf("error getting TX info for ClearDelegate: %w", err)
	}
	return nil
}
