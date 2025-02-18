package megapool

import (
	"errors"
	"fmt"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"

	"github.com/urfave/cli/v2"
)

func getStatus(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c)
	if err != nil {
		return err
	}

	// Get the config
	cfg, isNew, err := rp.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	// Get wallet status
	statusResponse, err := rp.Api.Wallet.Status()
	if err != nil {
		return err
	}
	walletStatus := statusResponse.Data.WalletStatus

	// Print what network we're on
	err = utils.PrintNetwork(cfg.Network.Value, isNew)
	if err != nil {
		return err
	}

	// Check if Saturn is deployed
	saturnResp, err := rp.Api.Network.IsSaturnDeployed()
	if err != nil {
		return err
	}
	if !saturnResp.Data.IsSaturnDeployed {
		fmt.Println("This command is only available after the Saturn upgrade.")
		return nil
	}

	// rp.Api.Megapool.Status() will fail with an error, but we can short-circuit it here.
	if !walletStatus.Address.HasAddress {
		return errors.New("Node Wallet is not initialized.")
	}

	fmt.Println("hello world")

	// Get megapool status
	status, err := rp.Api.Node.Status()
	if err != nil {
		return err
	}

	fmt.Println(status)

	return nil
}
