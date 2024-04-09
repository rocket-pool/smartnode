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
		router, "snapshot-delegate/clear", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeClearSnapshotDelegateContext struct {
	handler *NodeHandler
}

func (c *nodeClearSnapshotDelegateContext) PrepareData(data *types.TxInfoData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	cfg := sp.GetConfig()
	snapshot := sp.GetSnapshotDelegation()
	idHash := cfg.GetVotingSnapshotID()

	// Requirements
	err := sp.RequireNodeAddress()
	if err != nil {
		return types.ResponseStatus_AddressNotPresent, err
	}

	data.TxInfo, err = snapshot.ClearDelegate(idHash, opts)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for ClearDelegate: %w", err)
	}
	return types.ResponseStatus_Success, nil
}
