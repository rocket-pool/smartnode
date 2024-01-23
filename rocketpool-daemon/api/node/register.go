package node

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"

	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
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
		router, "register", f, f.handler.serviceProvider,
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

func (c *nodeRegisterContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	nodeAddress, _ := sp.GetWallet().GetAddress()

	// Requirements
	err := sp.RequireNodeRegistered()
	if err != nil {
		return err
	}

	// Bindings
	c.node, err = node.NewNode(c.rp, nodeAddress)
	if err != nil {
		return fmt.Errorf("error creating node %s binding: %w", nodeAddress.Hex(), err)
	}
	pMgr, err := protocol.NewProtocolDaoManager(c.rp)
	if err != nil {
		return fmt.Errorf("error getting pDAO manager binding: %w", err)
	}
	c.pSettings = pMgr.Settings
	return nil
}

func (c *nodeRegisterContext) GetState(mc *batch.MultiCaller) {
	core.AddQueryablesToMulticall(mc,
		c.node.Exists,
		c.pSettings.Node.IsRegistrationEnabled,
	)
}

func (c *nodeRegisterContext) PrepareData(data *api.NodeRegisterData, opts *bind.TransactOpts) error {
	data.AlreadyRegistered = c.node.Exists.Get()
	data.RegistrationDisabled = !c.pSettings.Node.IsRegistrationEnabled.Get()
	data.CanRegister = !(data.AlreadyRegistered || data.RegistrationDisabled)
	if !data.CanRegister {
		return nil
	}

	// Get tx info
	txInfo, err := c.node.Register(c.timezoneLocation, opts)
	if err != nil {
		return fmt.Errorf("error getting TX info for Register: %w", err)
	}
	data.TxInfo = txInfo
	return nil
}
