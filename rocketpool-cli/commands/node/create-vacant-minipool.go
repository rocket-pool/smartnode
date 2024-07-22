package node

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/commands/wallet"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/migration"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/terminal"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
	"github.com/urfave/cli/v2"
)

const (
	cvmMnemonicFlag string = "mnemonic"
)

func createVacantMinipool(c *cli.Context, pubkey beacon.ValidatorPubkey) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c)
	if err != nil {
		return err
	}

	d, err := newDepositPrompts(c, rp, &pubkey)
	if err != nil {
		return err
	}
	if d == nil {
		return nil
	}

	minNodeFee := d.minNodeFee
	amountWei := d.amountWei
	amount := eth.WeiToEth(amountWei)

	// Build the TX
	response, err := rp.Api.Node.CreateVacantMinipool(amountWei, minNodeFee, d.salt, pubkey)
	if err != nil {
		return err
	}

	// Verify
	if !response.Data.CanDeposit {
		fmt.Println("Cannot create a vacant minipool for migration:")
		if response.Data.InsufficientRplStake {
			fmt.Printf("The node has not staked enough RPL to collateralize a new minipool with a bond of %d ETH.\n", int(amount))
		}
		if response.Data.InvalidAmount {
			fmt.Println("The deposit amount is invalid.")
		}
		if response.Data.DepositDisabled {
			fmt.Println("Vacant minipool deposits are currently disabled.")
		}
		return nil
	}

	// Print the salt and minipool address info
	if c.String(saltFlag) != "" {
		fmt.Printf("Using custom salt %s, your minipool address will be %s.\n\n", c.String(saltFlag), response.Data.MinipoolAddress.Hex())
	}

	// Check to see if eth2 is synced
	syncResponse, err := rp.Api.Service.ClientStatus()
	if err != nil {
		return fmt.Errorf("error checking if your clients are in sync: %w", err)
	}
	if syncResponse.Data.BcManagerStatus.PrimaryClientStatus.IsSynced {
		fmt.Println("Your consensus client is synced, you may safely create a minipool.")
	} else if syncResponse.Data.BcManagerStatus.FallbackEnabled {
		if syncResponse.Data.BcManagerStatus.FallbackClientStatus.IsSynced {
			fmt.Println("Your fallback consensus client is synced, you may safely create a minipool.")
		} else {
			fmt.Print(terminal.ColorRed)
			fmt.Println("**WARNING**: neither your primary nor fallback consensus clients are fully synced.")
			fmt.Println("You cannot migrate until they've finished syncing.")
			fmt.Print(terminal.ColorReset)
			return nil
		}
	} else {
		fmt.Print(terminal.ColorRed)
		fmt.Println("**WARNING**: your primary consensus client is either not fully synced or offline and you do not have a fallback client configured.")
		fmt.Println("You cannot migrate until you have a synced consensus client.")
		fmt.Print(terminal.ColorReset)
		return nil
	}

	// Run the TX
	validated, err := tx.HandleTx(c, rp, response.Data.TxInfo,
		fmt.Sprintf(
			"You are about to create a new, vacant minipool with a minimum possible commission rate of %f%%. Once created, you will be able to migrate your existing validator into this minipool.\n"+
				"%sAre you sure you want to do this?%s",
			minNodeFee*100,
			terminal.ColorYellow,
			terminal.ColorReset),
		"creating minipool",
		"Creating minipool...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log
	fmt.Println("Your minipool was made successfully!")
	fmt.Printf("Your new minipool's address is: %s\n\n", response.Data.MinipoolAddress)

	// Get the mnemonic if importing
	mnemonic := ""
	if c.IsSet(cvmMnemonicFlag) {
		mnemonic = c.String(cvmMnemonicFlag)
	} else if !c.Bool(utils.YesFlag.Name) {
		fmt.Println("You have the option of importing your validator's private key into the Smartnode's Validator Client instead of running your own Validator Client separately. In doing so, the Smartnode will also automatically migrate your validator's withdrawal credentials from your BLS private key to the minipool you just created.")
		fmt.Println()
		if utils.Confirm("Would you like to import your key and automatically migrate your withdrawal credentials?") {
			mnemonic = wallet.PromptMnemonic()
		}
	}

	if mnemonic != "" {
		handleImport(c, rp, response.Data.MinipoolAddress, mnemonic)
	} else {
		// Ignore importing / it errored out
		fmt.Println("Since you're not importing your validator key, you will still be responsible for running and maintaining your own Validator Client with the validator's private key loaded, just as you are today.")
		fmt.Println()
		fmt.Println()
		fmt.Printf("You must now upgrade your validator's withdrawal credentials manually, using as tool such as `ethdo` (https://github.com/wealdtech/ethdo), to the following minipool address:\n\n\t%s\n\n", response.Data.MinipoolAddress)
	}

	fmt.Printf("The minipool is now in the scrub check, where it will hold for %s.\n", response.Data.ScrubPeriod)
	fmt.Println("You can watch its progress using `rocketpool service logs node`.")
	fmt.Println("Once the scrub check period has passed, your node will automatically promote it to an active minipool.")

	return nil

}

// Import a validator's private key into the Smartnode and set the validator's withdrawal creds
func handleImport(c *cli.Context, rp *client.Client, minipoolAddress common.Address, mnemonic string) {
	// Check if the withdrawal creds can be changed
	success := migration.ChangeWithdrawalCreds(rp, minipoolAddress, mnemonic)
	if !success {
		fmt.Println("Your withdrawal credentials cannot be automatically changed at this time. Import aborted.\nYou can try again later by using `rocketpool minipool set-withdrawal-creds`.")
		return
	}

	// Import the private key
	success = migration.ImportKey(c, rp, minipoolAddress, mnemonic)
	if !success {
		fmt.Println("Your validator's withdrawal credentials have been changed to the minipool address, but importing the key failed.\nYou can try again later by using `rocketpool minipool import-key`.")
		return
	}
}
