package service

import (
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/urfave/cli/v2"
)

// View the Rocket Pool service compose config
func serviceCompose(c *cli.Context) error {
	// Get RP client
	rp := client.NewClientFromCtx(c)

	// Print service compose config
	return rp.PrintServiceCompose(getComposeFiles(c))
}
