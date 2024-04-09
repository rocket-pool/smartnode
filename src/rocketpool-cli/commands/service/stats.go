package service

import (
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/urfave/cli/v2"
)

// View the Rocket Pool service stats
func serviceStats(c *cli.Context) error {
	// Get RP client
	rp := client.NewClientFromCtx(c)

	// Print service stats
	return rp.PrintServiceStats(getComposeFiles(c))
}
