package node

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/utils/input"
)

// ===============
// === Factory ===
// ===============

type nodeSetSnapshotDelegateContextFactory struct {
	handler *NodeHandler
}

func (f *nodeSetSnapshotDelegateContextFactory) Create(args url.Values) (*nodeSetSnapshotDelegateContext, error) {
	c := &nodeSetSnapshotDelegateContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("delegate", args, input.ValidateAddress, &c.delegate),
	}
	return c, errors.Join(inputErrs...)
}

func (f *nodeSetSnapshotDelegateContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*nodeSetSnapshotDelegateContext, types.TxInfoData](
		router, "snapshot-delegate/set", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeSetSnapshotDelegateContext struct {
	handler *NodeHandler

	delegate common.Address
}

func (c *nodeSetSnapshotDelegateContext) PrepareData(data *types.TxInfoData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	cfg := sp.GetConfig()
	snapshot := sp.GetSnapshotDelegation()
	idHash := cfg.GetVotingSnapshotID()

	// Requirements
	err := sp.RequireSnapshot()
	if err != nil {
		return types.ResponseStatus_InvalidChainState, err
	}

	data.TxInfo, err = snapshot.SetDelegate(idHash, c.delegate, opts)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for SetDelegate: %w", err)
	}
	return types.ResponseStatus_Success, nil
}
