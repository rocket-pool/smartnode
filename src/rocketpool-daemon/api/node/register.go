package node

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/rocketpool-go/v2/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/v2/node"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type nodeRegisterContextFactory struct {
	handler *NodeHandler
}

func (f *nodeRegisterContextFactory) Create(args url.Values) (*nodeRegisterContext, error) {
	c := &nodeRegisterContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("timezone", args, input.ValidateTimezoneLocation, &c.timezoneLocation),
	}
	return c, errors.Join(inputErrs...)
}

func (f *nodeRegisterContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*nodeRegisterContext, api.NodeRegisterData](
		router, "register", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeRegisterContext struct {
	handler *NodeHandler
	rp      *rocketpool.RocketPool

	timezoneLocation string
	node             *node.Node
	pSettings        *protocol.ProtocolDaoSettings
}

func (c *nodeRegisterContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	nodeAddress, _ := sp.GetWallet().GetAddress()

	// Requirements
	status, err := sp.RequireRocketPoolContracts(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	c.node, err = node.NewNode(c.rp, nodeAddress)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating node %s binding: %w", nodeAddress.Hex(), err)
	}
	pMgr, err := protocol.NewProtocolDaoManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting pDAO manager binding: %w", err)
	}
	c.pSettings = pMgr.Settings
	return types.ResponseStatus_Success, nil
}

func (c *nodeRegisterContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.node.Exists,
		c.pSettings.Node.IsRegistrationEnabled,
	)
}

func (c *nodeRegisterContext) PrepareData(data *api.NodeRegisterData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	data.AlreadyRegistered = c.node.Exists.Get()
	data.RegistrationDisabled = !c.pSettings.Node.IsRegistrationEnabled.Get()
	data.CanRegister = !(data.AlreadyRegistered || data.RegistrationDisabled)
	if !data.CanRegister {
		return types.ResponseStatus_Success, nil
	}

	// Get tx info
	txInfo, err := c.node.Register(c.timezoneLocation, opts)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for Register: %w", err)
	}
	data.TxInfo = txInfo
	return types.ResponseStatus_Success, nil
}
