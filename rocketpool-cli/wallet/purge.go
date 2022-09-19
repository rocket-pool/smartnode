package wallet

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/rocketpool-cli/service"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

func purge(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c)
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get the config
	cfg, _, err := rp.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading user settings: %w", err)
	}

	if !cliutils.Confirm(fmt.Sprintf("%sWARNING: This will delete your node wallet, all of your validator keys (including externally-generated ones in the 'custom-keys' folder), and restart your Validator Client.\nYou will NO LONGER be able to attest with this machine anymore until you recover your wallet or initialize a new one.\n\nYou MUST have your node wallet's mnemonic recorded before running this, or you will lose access to your node wallet and your validators forever!\n\n%sDo you want to continue?", colorRed, colorReset)) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Purge
	_, err = rp.Purge()
	if err != nil {
		return err
	}

	// Restart RP node and watchtower now that the wallet's gone
	if !cfg.IsNativeMode {
		projectName := cfg.Smartnode.ProjectName.Value.(string)
		nodeName := projectName + service.NodeContainerSuffix
		watchtowerName := projectName + service.WatchtowerContainerSuffix

		// Restart node
		err := restartContainer(rp, nodeName)
		if err != nil {
			return err
		}

		// Restart watchtower
		err = restartContainer(rp, watchtowerName)
		if err != nil {
			return err
		}
	} else {
		fmt.Printf("%sNOTE: As you are in Native mode, please restart your node and watchtower services manually to remove the cached wallet information.%s\n\n", colorYellow, colorReset)
	}

	fmt.Println("Deleted the node wallet and all validator keys.")
	return nil

}

func restartContainer(rp *rocketpool.Client, containerName string) error {
	// Restart node
	result, err := rp.RestartContainer(containerName)
	if err != nil {
		return fmt.Errorf("Error stopping %s container: %w", containerName, err)
	}
	if result != containerName {
		return fmt.Errorf("Unexpected output while stopping %s container: %s", containerName, result)
	}
	return nil
}
