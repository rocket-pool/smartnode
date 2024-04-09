package network

import (
	"fmt"
	"net/url"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/v2/node"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
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
		router, "timezone-map", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
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

func (c *networkTimezoneContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()

	// Requirements
	status, err := sp.RequireRocketPoolContracts(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	c.nodeMgr, err = node.NewNodeManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting node manager binding: %w", err)
	}
	return types.ResponseStatus_Success, nil
}

func (c *networkTimezoneContext) GetState(mc *batch.MultiCaller) {
	c.nodeMgr.NodeCount.AddToQuery(mc)
}

func (c *networkTimezoneContext) PrepareData(data *api.NetworkTimezonesData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	data.TimezoneCounts = map[string]uint64{}
	timezoneCounts, err := c.nodeMgr.GetNodeCountPerTimezone(c.nodeMgr.NodeCount.Formatted(), nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting node counts per timezone: %w", err)
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

	return types.ResponseStatus_Success, nil
}
