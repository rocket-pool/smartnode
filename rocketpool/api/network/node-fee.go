package network

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/rocketpool"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
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
		router, "node-fee", f, f.handler.serviceProvider,
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

func (c *networkFeeContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()

	// Requirements
	err := sp.RequireEthClientSynced()
	if err != nil {
		return err
	}

	// Bindings
	pMgr, err := protocol.NewProtocolDaoManager(c.rp)
	if err != nil {
		return fmt.Errorf("error creating pDAO manager binding: %w", err)
	}
	c.pSettings = pMgr.Settings
	c.networkMgr, err = network.NewNetworkManager(c.rp)
	if err != nil {
		return fmt.Errorf("error creating network manager binding: %w", err)
	}
	return nil
}

func (c *networkFeeContext) GetState(mc *batch.MultiCaller) {
	core.AddQueryablesToMulticall(mc,
		c.networkMgr.NodeFee,
		c.pSettings.Network.MinimumNodeFee,
		c.pSettings.Network.TargetNodeFee,
		c.pSettings.Network.MaximumNodeFee,
	)
}

func (c *networkFeeContext) PrepareData(data *api.NetworkNodeFeeData, opts *bind.TransactOpts) error {
	data.NodeFee = c.networkMgr.NodeFee.Formatted()
	data.MinNodeFee = c.pSettings.Network.MinimumNodeFee.Formatted()
	data.TargetNodeFee = c.pSettings.Network.TargetNodeFee.Formatted()
	data.MaxNodeFee = c.pSettings.Network.MaximumNodeFee.Formatted()
	return nil
}
