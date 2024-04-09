package network

import (
	"fmt"
	"sort"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/urfave/cli/v2"
)

func getTimezones(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get the timezone map
	response, err := rp.Api.Network.TimezoneMap()
	if err != nil {
		return err
	}

	// Sort it by the timezone name
	var maxNameLength int
	timezoneNames := make([]string, 0, len(response.Data.TimezoneCounts))
	for timezoneName := range response.Data.TimezoneCounts {
		if timezoneName != "Other" {
			timezoneNames = append(timezoneNames, timezoneName)
			nameLength := len(timezoneName) + 2
			if nameLength > maxNameLength {
				maxNameLength = nameLength
			}
		}
	}
	sort.Strings(timezoneNames)

	fmt.Printf("There are currently %d nodes across %d timezones.\n\n", response.Data.NodeTotal, response.Data.TimezoneTotal)

	for _, timezoneName := range timezoneNames {
		fmt.Printf("%-*s%d\n", maxNameLength, timezoneName+":", response.Data.TimezoneCounts[timezoneName])
	}
	fmt.Printf("%-*s%d\n", maxNameLength, "Other:", response.Data.TimezoneCounts["Other"])

	return nil
}
