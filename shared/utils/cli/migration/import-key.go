package migration

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/rocketpool-cli/wallet"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/utils/cli/color"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/urfave/cli"
)

// Imports the private key for a vacant minipool's validator
func ImportKey(c *cli.Context, rp *rocketpool.Client, minipoolAddress common.Address, mnemonic string) bool {

	// Print a warning and prompt for confirmation of anti-slashing
	color.RedPrintln("WARNING:")
	color.RedPrintln("Before doing this, you **MUST** do the following:")
	color.RedPrintln("1. Remove this key from your existing Validator Client used for solo staking")
	color.RedPrintln("2. Restart it so that it is no longer validating with that key")
	color.RedPrintln("3. Wait for 15 minutes so it has missed at least two attestations")
	color.RedPrintln("Failure to do this **will result in your validator being SLASHED**.")
	fmt.Println()
	if !prompt.Confirm("Have you removed the key from your own Validator Client, restarted it, and waited long enough for your validator to miss at least two attestations?") {
		fmt.Println("Cancelled.")
		return false
	}

	// Get the mnemonic
	if mnemonic == "" {
		mnemonic = wallet.PromptMnemonic()
	}

	// Import the key
	fmt.Printf("Importing validator key... ")
	_, err := rp.ImportKey(minipoolAddress, mnemonic)
	if err != nil {
		fmt.Printf("error importing validator key: %s\n", err.Error())
		return false
	}
	fmt.Println("done!")

	// Restart the VC if necessary
	if c.Bool("no-restart") {
		return true
	}
	if c.Bool("yes") || prompt.Confirm("Would you like to restart the Smart Node's Validator Client now so it loads your validator's key?") {
		// Restart the VC
		fmt.Print("Restarting Validator Client... ")
		_, err := rp.RestartVc()
		if err != nil {
			fmt.Println("failed!")
			color.YellowPrintf("WARNING: error restarting validator client: %s\n", err.Error())
			fmt.Println()
			color.YellowPrintln("Please restart it manually so it picks up the new validator key for your minipool.")
			return false
		}
		fmt.Println("done!")
		fmt.Println()
	}
	return true

}
