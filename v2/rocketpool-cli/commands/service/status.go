package service

import (
	"fmt"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/urfave/cli/v2"
)

// View the Rocket Pool service status
func serviceStatus(c *cli.Context) error {
	// Get RP client
	rp := client.NewClientFromCtx(c)

	// Get the config
	cfg, isNew, err := rp.LoadConfig()
	if err != nil {
		return fmt.Errorf("Error loading configuration: %w", err)
	}

	// Print what network we're on
	err = utils.PrintNetwork(cfg.Network.Value, isNew)
	if err != nil {
		return err
	}

	// Print service status
	return rp.PrintServiceStatus(getComposeFiles(c))
}
