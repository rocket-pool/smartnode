package network

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/v2/network"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type networkFeeContextFactory struct {
	handler *NetworkHandler
}

func (f *networkFeeContextFactory) Create(args url.Values) (*networkFeeContext, error) {
	c := &networkFeeContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *networkFeeContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*networkFeeContext, api.NetworkNodeFeeData](
		router, "node-fee", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type networkFeeContext struct {
	handler *NetworkHandler
	rp      *rocketpool.RocketPool

	pSettings  *protocol.ProtocolDaoSettings
	networkMgr *network.NetworkManager
}

func (c *networkFeeContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()

	// Requirements
	status, err := sp.RequireRocketPoolContracts(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	pMgr, err := protocol.NewProtocolDaoManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating pDAO manager binding: %w", err)
	}
	c.pSettings = pMgr.Settings
	c.networkMgr, err = network.NewNetworkManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating network manager binding: %w", err)
	}
	return types.ResponseStatus_Success, nil
}

func (c *networkFeeContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.networkMgr.NodeFee,
		c.pSettings.Network.MinimumNodeFee,
		c.pSettings.Network.TargetNodeFee,
		c.pSettings.Network.MaximumNodeFee,
	)
}

func (c *networkFeeContext) PrepareData(data *api.NetworkNodeFeeData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	data.NodeFee = c.networkMgr.NodeFee.Raw()
	data.MinNodeFee = c.pSettings.Network.MinimumNodeFee.Raw()
	data.TargetNodeFee = c.pSettings.Network.TargetNodeFee.Raw()
	data.MaxNodeFee = c.pSettings.Network.MaximumNodeFee.Raw()
	return types.ResponseStatus_Success, nil
}
