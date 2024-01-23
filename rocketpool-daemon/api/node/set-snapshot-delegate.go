package node

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"

	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
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
	server.RegisterQuerylessGet[*nodeSetSnapshotDelegateContext, api.TxInfoData](
		router, "snapshot-delegate/set", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeSetSnapshotDelegateContext struct {
	handler *NodeHandler

	delegate common.Address
}

func (c *nodeSetSnapshotDelegateContext) PrepareData(data *api.TxInfoData, opts *bind.TransactOpts) error {
	sp := c.handler.serviceProvider
	cfg := sp.GetConfig()
	snapshot := sp.GetSnapshotDelegation()
	idHash := cfg.Smartnode.GetVotingSnapshotID()

	var err error
	data.TxInfo, err = snapshot.SetDelegate(idHash, c.delegate, opts)
	if err != nil {
		return fmt.Errorf("error getting TX info for SetDelegate: %w", err)
	}
	return nil
}
