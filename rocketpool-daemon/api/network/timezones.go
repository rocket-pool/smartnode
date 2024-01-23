package network

import (
	"fmt"
	"net/url"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"

	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type networkTimezoneContextFactory struct {
	handler *NetworkHandler
}

func (f *networkTimezoneContextFactory) Create(args url.Values) (*networkTimezoneContext, error) {
	c := &networkTimezoneContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *networkTimezoneContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*networkTimezoneContext, api.NetworkTimezonesData](
		router, "timezone-map", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type networkTimezoneContext struct {
	handler *NetworkHandler
	rp      *rocketpool.RocketPool

	nodeMgr *node.NodeManager
}

func (c *networkTimezoneContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()

	// Requirements
	err := sp.RequireEthClientSynced()
	if err != nil {
		return err
	}

	// Bindings
	c.nodeMgr, err = node.NewNodeManager(c.rp)
	if err != nil {
		return fmt.Errorf("error getting node manager binding: %w", err)
	}
	return nil
}

func (c *networkTimezoneContext) GetState(mc *batch.MultiCaller) {
	c.nodeMgr.NodeCount.AddToQuery(mc)
}

func (c *networkTimezoneContext) PrepareData(data *api.NetworkTimezonesData, opts *bind.TransactOpts) error {
	data.TimezoneCounts = map[string]uint64{}
	timezoneCounts, err := c.nodeMgr.GetNodeCountPerTimezone(c.nodeMgr.NodeCount.Formatted(), nil)
	if err != nil {
		return fmt.Errorf("error getting node counts per timezone: %w", err)
	}

	for timezone, count := range timezoneCounts {
		location, err := time.LoadLocation(timezone)
		if err != nil {
			data.TimezoneCounts["Other"] += count
		} else {
			data.TimezoneCounts[location.String()] = count
		}
		data.TimezoneTotal++
		data.NodeTotal += count
	}

	return nil
}
