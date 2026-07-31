package wallet

import (
	"fmt"

	cliutils "github.com/rocket-pool/smartnode/rocketpool-cli/cli"
	"github.com/rocket-pool/smartnode/rocketpool-cli/cli/color"
	"github.com/rocket-pool/smartnode/rocketpool-cli/cli/prompt"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
)

func masquerade(addressFlag string, yes bool, observe bool) error {
	// Get RP client
	rp := rocketpool.NewClient()
	defer rp.Close()

	fmt.Println("Masquerading allows you to set your node address to any address you want. All commands will act as though your node wallet is for that address. Since you don't have the private key for that address, you can't submit transactions or sign messages though; commands will be", color.Yellow("read-only"), "until you end the masquerade with `rocketpool wallet end-masquerade`.")
	fmt.Println()

	// Get address
	if addressFlag == "" {
		addressFlag = prompt.Prompt("Please enter an address to masquerade as:", "^0x[0-9a-fA-F]{40}$", "Invalid address")
	}

	address, err := cliutils.ValidateAddress("address", addressFlag)
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if observe {
		fmt.Println(color.Yellow("Observe mode is enabled. Please read the following carefully:"))
		fmt.Println(" - The node and watchtower will use the masquerade address instead of your real node address.")
		fmt.Println(" - Your fee recipient will remain set to your real node wallet address")
		fmt.Println(" - Run `rocketpool wallet end-masquerade` and restart the node/watchtower daemons when you have finished observing.")
		fmt.Println()
		if !yes && !prompt.Confirm("I understand the above. Enable observe mode for %s?", color.LightBlue(address.Hex())) {
			fmt.Println("Cancelled.")
			return nil
		}
	} else {
		if !yes && !prompt.Confirm("Are you sure you want to masquerade as %s?", color.LightBlue(address.Hex())) {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// Call API
	_, err = rp.Masquerade(address, observe)
	if err != nil {
		return fmt.Errorf("error running masquerade: %w", err)
	}

	fmt.Printf("Your node is now masquerading as address %s.\n", color.LightBlue(address.Hex()))
	if observe {
		return promptAndRestartObserveDaemons(rp, yes)
	}

	return nil
}

// promptAndRestartObserveDaemons asks the user whether to restart the node and watchtower
// containers now, so they immediately pick up the new observe-mode state.
func promptAndRestartObserveDaemons(rp *rocketpool.Client, yes bool) error {
	cfg, _, err := rp.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	if cfg.IsNativeMode {
		fmt.Println("Restart the node and watchtower daemons for this to take effect.")
		return nil
	}

	if !yes && !prompt.Confirm("Would you like to restart the node and watchtower containers now?") {
		fmt.Println("Remember to restart the node and watchtower containers for this to take effect.")
		return nil
	}

	projectName := cfg.Smartnode.ProjectName.Value.(string)
	for _, name := range []string{"node", "watchtower"} {
		container := fmt.Sprintf("%s_%s", projectName, name)
		fmt.Printf("Restarting %s... ", container)
		response, err := rp.RestartContainer(container)
		if err != nil {
			fmt.Println()
			return fmt.Errorf("error restarting %s: %w", container, err)
		}
		if response != container {
			fmt.Println()
			return fmt.Errorf("unexpected output while restarting %s: %s", container, response)
		}
		fmt.Println("done!")
	}

	fmt.Println("Done!")
	return nil
}
