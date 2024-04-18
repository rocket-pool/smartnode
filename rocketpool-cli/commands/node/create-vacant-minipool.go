package node

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"

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
	rp := client.NewClientFromCtx(c)

	// Make sure Beacon is on the correct chain
	depositContractInfo, err := rp.Api.Network.GetDepositContractInfo()
	if err != nil {
		return err
	}
	if depositContractInfo.Data.RPNetwork != depositContractInfo.Data.BeaconNetwork ||
		depositContractInfo.Data.RPDepositContract != depositContractInfo.Data.BeaconDepositContract {
		utils.PrintDepositMismatchError(
			depositContractInfo.Data.RPNetwork,
			depositContractInfo.Data.BeaconNetwork,
			depositContractInfo.Data.RPDepositContract,
			depositContractInfo.Data.BeaconDepositContract)
		return nil
	}

	fmt.Println("Your Beacon Node is on the correct network.")
	fmt.Println()

	// Check if the fee distributor has been initialized
	feeDistributorResponse, err := rp.Api.Node.InitializeFeeDistributor()
	if err != nil {
		return err
	}
	if !feeDistributorResponse.Data.IsInitialized {
		fmt.Println("Your fee distributor has not been initialized yet so you cannot create a new minipool.\nPlease run `rocketpool node initialize-fee-distributor` to initialize it first.")
		return nil
	}

	// Post a warning about fee distribution
	if !(c.Bool("yes") || utils.Confirm(fmt.Sprintf("%sNOTE: by creating a new minipool, your node will automatically claim and distribute any balance you have in your fee distributor contract. If you don't want to claim your balance at this time, you should not create a new minipool.%s\nWould you like to continue?", terminal.ColorYellow, terminal.ColorReset))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Print a notification about the pubkey
	fmt.Printf("You are about to convert the solo staker %s into a Rocket Pool minipool. This will convert your 32 ETH deposit into either an 8 ETH or 16 ETH deposit (your choice), and convert the remaining 24 or 16 ETH into a deposit from the Rocket Pool staking pool. The staking pool portion will be credited to your node's account, allowing you to create more validators without depositing additional ETH onto the Beacon Chain. Your excess balance (your existing Beacon rewards) will be preserved and not shared with the pool stakers.\n\nPlease thoroughly read our documentation at https://docs.rocketpool.net/guides/atlas/solo-staker-migration.html to learn about the process and its implications.\n\n1. First, we'll create the new minipool.\n2. Next, we'll ask whether you want to import the validator's private key into your Smartnode's Validator Client, or keep running your own externally-managed validator.\n3. Finally, we'll help you migrate your validator's withdrawal credentials to the minipool address.\n\n%sNOTE: If you intend to use the credit balance to create additional validators, you will need to have enough RPL staked to support them.%s\n\n", pubkey.Hex(), terminal.ColorYellow, terminal.ColorReset)

	// Get deposit amount
	var amount float64
	if c.String(amountFlag) != "" {
		// Parse amount
		depositAmount, err := strconv.ParseFloat(c.String(amountFlag), 64)
		if err != nil {
			return fmt.Errorf("invalid deposit amount '%s': %w", c.String(amountFlag), err)
		}
		amount = depositAmount
	} else {
		// Get deposit amount options
		amountOptions := []string{
			"8 ETH",
			"16 ETH",
		}

		// Prompt for amount
		selected, _ := utils.Select("Please choose an amount of ETH you want to use as your deposit for the new minipool (this will become your share of the balance, and the remainder will become the pool stakers' share):", amountOptions)
		switch selected {
		case 0:
			amount = 8
		case 1:
			amount = 16
		}
	}
	amountWei := eth.EthToWei(amount)

	// Get network node fees
	nodeFeeResponse, err := rp.Api.Network.NodeFee()
	if err != nil {
		return err
	}

	// Get minimum node fee
	var minNodeFee float64
	if c.String("max-slippage") == "auto" {
		// Use default max slippage
		minNodeFee = eth.WeiToEth(nodeFeeResponse.Data.NodeFee) - DefaultMaxNodeFeeSlippage
		if minNodeFee < eth.WeiToEth(nodeFeeResponse.Data.MinNodeFee) {
			minNodeFee = eth.WeiToEth(nodeFeeResponse.Data.MinNodeFee)
		}
	} else if c.String("max-slippage") != "" {
		// Parse max slippage
		maxNodeFeeSlippagePerc, err := strconv.ParseFloat(c.String("max-slippage"), 64)
		if err != nil {
			return fmt.Errorf("invalid maximum commission rate slippage '%s': %w", c.String("max-slippage"), err)
		}
		maxNodeFeeSlippage := maxNodeFeeSlippagePerc / 100

		// Calculate min node fee
		minNodeFee = eth.WeiToEth(nodeFeeResponse.Data.NodeFee) - maxNodeFeeSlippage
		if minNodeFee < eth.WeiToEth(nodeFeeResponse.Data.MinNodeFee) {
			minNodeFee = eth.WeiToEth(nodeFeeResponse.Data.MinNodeFee)
		}
	} else {
		// Prompt for min node fee
		if nodeFeeResponse.Data.MinNodeFee == nodeFeeResponse.Data.MaxNodeFee {
			fmt.Printf("Your minipool will use the current fixed commission rate of %.2f%%.\n", eth.WeiToEth(nodeFeeResponse.Data.MinNodeFee)*100)
			minNodeFee = eth.WeiToEth(nodeFeeResponse.Data.MinNodeFee)
		} else {
			minNodeFee = promptMinNodeFee(eth.WeiToEth(nodeFeeResponse.Data.NodeFee), eth.WeiToEth(nodeFeeResponse.Data.MinNodeFee))
		}
	}

	// Get minipool salt
	var salt *big.Int
	if c.String(saltFlag) != "" {
		var success bool
		salt, success = big.NewInt(0).SetString(c.String(saltFlag), 0)
		if !success {
			return fmt.Errorf("invalid minipool salt: %s", c.String(saltFlag))
		}
	} else {
		buffer := make([]byte, 32)
		_, err = rand.Read(buffer)
		if err != nil {
			return fmt.Errorf("error generating random salt: %w", err)
		}
		salt = big.NewInt(0).SetBytes(buffer)
	}

	// Build the TX
	response, err := rp.Api.Node.CreateVacantMinipool(amountWei, minNodeFee, salt, pubkey)
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
	} else {
		if syncResponse.Data.BcManagerStatus.PrimaryClientStatus.IsSynced {
			fmt.Printf("Your consensus client is synced, you may safely create a minipool.\n")
		} else if syncResponse.Data.BcManagerStatus.FallbackEnabled {
			if syncResponse.Data.BcManagerStatus.FallbackClientStatus.IsSynced {
				fmt.Printf("Your fallback consensus client is synced, you may safely create a minipool.\n")
			} else {
				fmt.Printf("%s**WARNING**: neither your primary nor fallback consensus clients are fully synced.\nYou cannot migrate until they've finished syncing.\n%s", terminal.ColorRed, terminal.ColorReset)
				return nil
			}
		} else {
			fmt.Printf("%s**WARNING**: your primary consensus client is either not fully synced or offline and you do not have a fallback client configured.\nYou cannot migrate until you have a synced consensus client.\n%s", terminal.ColorRed, terminal.ColorReset)
			return nil
		}
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
