package pdao

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
	"github.com/rocket-pool/rocketpool-go/v2/node"
)

// ===============
// === Factory ===
// ===============

type protocolDaoSetVotingDelegateContextFactory struct {
	handler *ProtocolDaoHandler
}

func (f *protocolDaoSetVotingDelegateContextFactory) Create(args url.Values) (*protocolDaoSetSnapshotDelegateContext, error) {
	c := &protocolDaoSetSnapshotDelegateContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("delegate", args, input.ValidateAddress, &c.delegate),
	}
	return c, errors.Join(inputErrs...)
}

func (f *protocolDaoSetVotingDelegateContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*protocolDaoSetSnapshotDelegateContext, types.TxInfoData](
		router, "voting-delegate/set", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type protocolDaoSetSnapshotDelegateContext struct {
	handler *ProtocolDaoHandler

	delegate common.Address
}

func (c *protocolDaoSetSnapshotDelegateContext) PrepareData(data *types.TxInfoData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	rp := sp.GetRocketPool()
	nodeAddress, _ := sp.GetWallet().GetAddress()

	// Requirements
	status, err := sp.RequireNodeRegistered(c.handler.ctx)
	if err != nil {
		return status, nil
	}

	// Binding
	node, err := node.NewNode(rp, nodeAddress)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating node %s binding: %w", nodeAddress.Hex(), err)
	}

	// Get TX info
	data.TxInfo, err = node.SetVotingDelegate(c.delegate, opts)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for SetVotingDelegate: %w", err)
	}
	return types.ResponseStatus_Success, nil
}
