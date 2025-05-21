package network

import (
	"math/big"
	"time"

	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getTimezones(c *cli.Context) (*api.NetworkTimezonesResponse, error) {

	// Get services
	if err := services.RequireRocketStorage(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.NetworkTimezonesResponse{}
	response.TimezoneCounts = map[string]uint64{}

	zero := big.NewInt(0)
	timezoneCounts, err := node.GetNodeCountPerTimezone(rp, zero, zero, nil)
	if err != nil {
		return nil, err
	}

	for _, timezoneCount := range timezoneCounts {
		location, err := time.LoadLocation(timezoneCount.Timezone)
		count := timezoneCount.Count.Uint64()
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
