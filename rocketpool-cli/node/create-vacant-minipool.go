package node

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/rocket-pool/smartnode/rocketpool-cli/wallet"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/migration"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/urfave/cli"
)

func createVacantMinipool(c *cli.Context, pubkey types.ValidatorPubkey) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check if Saturn is already deployed
	saturnDeployed, err := rp.IsSaturnDeployed()
	if err != nil {
		return err
	}
	if saturnDeployed.IsSaturnDeployed {
		fmt.Println("You cannot create a vacant minipool because Saturn is already deployed.")
		return nil
	}

	// Make sure ETH2 is on the correct chain
	depositContractInfo, err := rp.DepositContractInfo()
	if err != nil {
		return err
	}
	if depositContractInfo.RPNetwork != depositContractInfo.BeaconNetwork ||
		depositContractInfo.RPDepositContract != depositContractInfo.BeaconDepositContract {
		cliutils.PrintDepositMismatchError(
			depositContractInfo.RPNetwork,
			depositContractInfo.BeaconNetwork,
			depositContractInfo.RPDepositContract,
			depositContractInfo.BeaconDepositContract)
		return nil
	}

	fmt.Println("Your eth2 client is on the correct network.")
	fmt.Println()

	// Check if the fee distributor has been initialized
	isInitializedResponse, err := rp.IsFeeDistributorInitialized()
	if err != nil {
		return err
	}
	if !isInitializedResponse.IsInitialized {
		fmt.Println("Your fee distributor has not been initialized yet so you cannot create a new minipool.\nPlease run `rocketpool node initialize-fee-distributor` to initialize it first.")
		return nil
	}

	// Post a warning about fee distribution
	if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("%sNOTE: By creating a new minipool, your node will automatically claim and distribute any balance you have in your fee distributor contract. If you don't want to claim your balance at this time, you should not create a new minipool.%s\nWould you like to continue?", colorYellow, colorReset))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Print a notification about the pubkey
	fmt.Printf("You are about to convert the solo staker %s into a Rocket Pool minipool. This will convert your 32 ETH deposit into an 8 ETH deposit, and convert the remaining 24 ETH into a deposit from the Rocket Pool staking pool. The staking pool portion will be credited to your node's account, allowing you to create more validators without depositing additional ETH onto the Beacon Chain. Your excess balance (your existing Beacon rewards) will be preserved and not shared with the pool stakers.\n\nPlease thoroughly read our documentation at https://docs.rocketpool.net/guides/atlas/solo-staker-migration.html to learn about the process and its implications.\n\n1. First, we'll create the new minipool.\n2. Next, we'll ask whether you want to import the validator's private key into your Smartnode's Validator Client, or keep running your own externally-managed validator.\n3. Finally, we'll help you migrate your validator's withdrawal credentials to the minipool address.\n\n%sNOTE: If you intend to use the credit balance to create additional validators, you will need to have enough RPL staked to support them.%s\n\n", pubkey.Hex(), colorYellow, colorReset)

	// Get deposit amount
	var amount float64
	if c.String("amount") != "" {

		// Parse amount
		depositAmount, err := strconv.ParseFloat(c.String("amount"), 64)
		if err != nil {
			return fmt.Errorf("Invalid deposit amount '%s': %w", c.String("amount"), err)
		}
		amount = depositAmount

	} else {
		if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("%sNOTE: Your new minipool will use an 8 ETH deposit (this will become your share of the balance, and the remainder will become the pool stakers' share):%s\nWould you like to continue?", colorYellow, colorReset))) {
			fmt.Println("Cancelled.")
			return nil
		}
		amount = 8
	}

	amountWei := eth.EthToWei(amount)

	// Get network node fees
	nodeFees, err := rp.NodeFee()
	if err != nil {
		return err
	}

	// Get minimum node fee
	var minNodeFee float64
	if c.String("max-slippage") == "auto" {

		// Use default max slippage
		minNodeFee = nodeFees.NodeFee - defaultMaxNodeFeeSlippage
		if minNodeFee < nodeFees.MinNodeFee {
			minNodeFee = nodeFees.MinNodeFee
		}

	} else if c.String("max-slippage") != "" {

		// Parse max slippage
		maxNodeFeeSlippagePerc, err := strconv.ParseFloat(c.String("max-slippage"), 64)
		if err != nil {
			return fmt.Errorf("Invalid maximum commission rate slippage '%s': %w", c.String("max-slippage"), err)
		}
		maxNodeFeeSlippage := maxNodeFeeSlippagePerc / 100

		// Calculate min node fee
		minNodeFee = nodeFees.NodeFee - maxNodeFeeSlippage
		if minNodeFee < nodeFees.MinNodeFee {
			minNodeFee = nodeFees.MinNodeFee
		}

	} else {

		// Prompt for min node fee
		if nodeFees.MinNodeFee == nodeFees.MaxNodeFee {
			fmt.Printf("Your minipool will use the current base commission rate of %.2f%%.\n", nodeFees.MinNodeFee*100)
			minNodeFee = nodeFees.MinNodeFee
		} else {
			minNodeFee = promptMinNodeFee(nodeFees.NodeFee, nodeFees.MinNodeFee)
		}

	}

	// Get minipool salt
	var salt *big.Int
	if c.String("salt") != "" {
		var success bool
		salt, success = big.NewInt(0).SetString(c.String("salt"), 0)
		if !success {
			return fmt.Errorf("Invalid minipool salt: %s", c.String("salt"))
		}
	} else {
		buffer := make([]byte, 32)
		_, err = rand.Read(buffer)
		if err != nil {
			return fmt.Errorf("Error generating random salt: %w", err)
		}
		salt = big.NewInt(0).SetBytes(buffer)
	}

	// Check deposit can be made
	canDeposit, err := rp.CanCreateVacantMinipool(amountWei, minNodeFee, salt, pubkey)
	if err != nil {
		return err
	}
	if !canDeposit.CanDeposit {
		fmt.Println("Cannot create a vacant minipool for migration:")
		if canDeposit.InsufficientRplStake {
			fmt.Printf("The node has not staked enough RPL to collateralize a new minipool with a bond of %d ETH.\n", int(amount))
		}
		if canDeposit.InvalidAmount {
			fmt.Println("The deposit amount is invalid.")
		}
		if canDeposit.DepositDisabled {
			fmt.Println("Vacant minipool deposits are currently disabled.")
		}
		return nil
	}

	if c.String("salt") != "" {
		fmt.Printf("Using custom salt %s, your minipool address will be %s.\n\n", c.String("salt"), canDeposit.MinipoolAddress.Hex())
	}

	// Check to see if eth2 is synced
	colorReset := "\033[0m"
	colorRed := "\033[31m"
	colorYellow := "\033[33m"
	syncResponse, err := rp.NodeSync()
	if err != nil {
		return fmt.Errorf("error checking if your clients are in sync: %w", err)
	} else {
		if syncResponse.BcStatus.PrimaryClientStatus.IsSynced {
			fmt.Printf("Your consensus client is synced, you may safely create a minipool.\n")
		} else if syncResponse.BcStatus.FallbackEnabled {
			if syncResponse.BcStatus.FallbackClientStatus.IsSynced {
				fmt.Printf("Your fallback consensus client is synced, you may safely create a minipool.\n")
			} else {
				fmt.Printf("%s**WARNING**: neither your primary nor fallback consensus clients are fully synced.\nYou cannot migrate until they've finished syncing.\n%s", colorRed, colorReset)
				return nil
			}
		} else {
			fmt.Printf("%s**WARNING**: your primary consensus client is either not fully synced or offline and you do not have a fallback client configured.\nYou cannot migrate until you have a synced consensus client.\n%s", colorRed, colorReset)
			return nil
		}
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canDeposit.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf(
		"You are about to create a new, vacant minipool with a minimum possible commission rate of %f%%. Once created, you will be able to migrate your existing validator into this minipool.\n"+
			"%sAre you sure you want to do this?%s",
		minNodeFee*100,
		colorYellow,
		colorReset))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Make deposit
	response, err := rp.CreateVacantMinipool(amountWei, minNodeFee, salt, pubkey)
	if err != nil {
		return err
	}

	// Log and wait for the minipool address
	fmt.Printf("Creating minipool...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	_, err = rp.WaitForTransaction(response.TxHash)
	if err != nil {
		return err
	}

	// Log
	fmt.Println("Your minipool was made successfully!")
	fmt.Printf("Your new minipool's address is: %s\n\n", response.MinipoolAddress)

	// Get the mnemonic if importing
	mnemonic := ""
	if c.IsSet("mnemonic") {
		mnemonic = c.String("mnemonic")
	} else if !c.Bool("yes") {
		fmt.Println("You have the option of importing your validator's private key into the Smartnode's Validator Client instead of running your own Validator Client separately. In doing so, the Smartnode will also automatically migrate your validator's withdrawal credentials from your BLS private key to the minipool you just created.")
		fmt.Println()
		if prompt.Confirm("Would you like to import your key and automatically migrate your withdrawal credentials?") {
			mnemonic = wallet.PromptMnemonic()
		}
	}

	if mnemonic != "" {
		handleImport(c, rp, response.MinipoolAddress, mnemonic)
	} else {
		// Ignore importing / it errored out
		fmt.Println("Since you're not importing your validator key, you will still be responsible for running and maintaining your own Validator Client with the validator's private key loaded, just as you are today.")
		fmt.Println()
		fmt.Println()
		fmt.Printf("You must now upgrade your validator's withdrawal credentials manually, using as tool such as `ethdo` (https://github.com/wealdtech/ethdo), to the following minipool address:\n\n\t%s\n\n", response.MinipoolAddress)
	}

	fmt.Printf("The minipool is now in the scrub check, where it will hold for %s.\n", response.ScrubPeriod)
	fmt.Println("You can watch its progress using `rocketpool service logs node`.")
	fmt.Println("Once the scrub check period has passed, your node will automatically promote it to an active minipool.")

	return nil

}

// Import a validator's private key into the Smartnode and set the validator's withdrawal creds
func handleImport(c *cli.Context, rp *rocketpool.Client, minipoolAddress common.Address, mnemonic string) {
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
