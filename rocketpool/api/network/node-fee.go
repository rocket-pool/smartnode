package network

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/settings"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type networkFeeContextFactory struct {
	handler *NetworkHandler
}

func (f *networkFeeContextFactory) Create(vars map[string]string) (*networkFeeContext, error) {
	c := &networkFeeContext{
		handler: f.handler,
	}
	return c, nil
}

// ===============
// === Context ===
// ===============

type networkFeeContext struct {
	handler *NetworkHandler
	rp      *rocketpool.RocketPool

	networkFees *network.NetworkFees
	pSettings   *settings.ProtocolDaoSettings
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
	c.networkFees, err = network.NewNetworkFees(c.rp)
	if err != nil {
		return fmt.Errorf("error getting network fees binding: %w", err)
	}
	c.pSettings, err = settings.NewProtocolDaoSettings(c.rp)
	if err != nil {
		return fmt.Errorf("error getting protocol DAO settings binding: %w", err)
	}
	return nil
}

func (c *networkFeeContext) GetState(mc *batch.MultiCaller) {
	c.networkFees.GetNodeFee(mc)
	c.pSettings.GetMinimumNodeFee(mc)
	c.pSettings.GetTargetNodeFee(mc)
	c.pSettings.GetMaximumNodeFee(mc)
}

func (c *networkFeeContext) PrepareData(data *api.NetworkNodeFeeData, opts *bind.TransactOpts) error {
	data.NodeFee = c.networkFees.Details.NodeFee.Formatted()
	data.MinNodeFee = c.pSettings.Details.Network.MinimumNodeFee.Formatted()
	data.TargetNodeFee = c.pSettings.Details.Network.TargetNodeFee.Formatted()
	data.MaxNodeFee = c.pSettings.Details.Network.MaximumNodeFee.Formatted()
	return nil
}
