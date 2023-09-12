package network

import (
	"fmt"
	"time"

	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getTimezones(c *cli.Context) (*api.NetworkTimezonesResponse, error) {
	// Get services
	if err := services.RequireEthClientSynced(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.NetworkTimezonesResponse{}
	response.TimezoneCounts = map[string]uint64{}

	// Create bindings
	nodeMgr, err := node.NewNodeManager(rp)
	if err != nil {
		return nil, fmt.Errorf("error getting node manager binding: %w", err)
	}

	// Get contract state
	err = rp.Query(func(mc *batch.MultiCaller) error {
		nodeMgr.GetNodeCount(mc)
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting contract state: %w", err)
	}

	timezoneCounts, err := nodeMgr.GetNodeCountPerTimezone(nodeMgr.Details.NodeCount.Formatted(), nil)
	if err != nil {
		return nil, fmt.Errorf("error getting node counts per timezone: %w", err)
	}

	for timezone, count := range timezoneCounts {
		location, err := time.LoadLocation(timezone)
		if err != nil {
			response.TimezoneCounts["Other"] += count
		} else {
			response.TimezoneCounts[location.String()] = count
		}
		response.TimezoneTotal++
		response.NodeTotal += count
	}

	// Return response
	return &response, nil
}
