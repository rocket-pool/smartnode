package node

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"

	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/math"
	"github.com/urfave/cli"
)

func createVacantMinipool(c *cli.Context, pubkey types.ValidatorPubkey) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c)
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check and assign the EC status
	err = cliutils.CheckClientStatus(rp)
	if err != nil {
		return err
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

	fmt.Println("Your eth2 client is on the correct network.\n")

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
	if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("%sNOTE: by creating a new minipool, your node will automatically claim and distribute any balance you have in your fee distributor contract. If you don't want to claim your balance at this time, you should not create a new minipool.%s\nWould you like to continue?", colorYellow, colorReset))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Print a notification about the pubkey
	fmt.Printf("You are about to convert the solo staker %s into a Rocket Pool minipool. This will convert your 32 ETH deposit into either an 8 ETH or 16 ETH deposit (your choice), and convert the remaining 24 or 16 ETH into a deposit from the Rocket Pool staking pool. The staking pool portion will be credited to your node's account, allowing you to create more validators without depositing additional ETH onto the Beacon Chain. Your excess balance (your existing Beacon rewards) will be preserved and not shared with the pool stakers.\n\nPlease thoroughly read our documentation at <placeholder> to learn about the process and its implications.\n\n", pubkey.Hex())

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

		// Get deposit amount options
		amountOptions := []string{
			"8 ETH",
			"16 ETH",
		}

		// Prompt for amount
		selected, _ := cliutils.Select("Please choose an amount of ETH you want to use as your deposit for the new minipool (this will become your share of the balance, and the remainder will become the pool stakers' share):", amountOptions)
		switch selected {
		case 0:
			amount = 8
		case 1:
			amount = 16
		}

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
		minNodeFee = nodeFees.NodeFee - DefaultMaxNodeFeeSlippage
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
			fmt.Printf("Your minipool will use the current fixed commission rate of %.2f%%.\n", nodeFees.MinNodeFee*100)
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

	// Ask amount importing
	importKey := c.Bool("import-key")
	if importKey {
		fmt.Printf("%sNOTE:\nYou have requested to import your validator's private key into the Validator Client container managed by the Smartnode stack. Before doing this, you **MUST** remove this key from your existing Validator Client used for solo staking and restart it so that it is no longer validating with that key.\nFailure to do this **will result in your validator being SLASHED**.\n\nPlease confirm the validator is no longer active by intentionally waiting until it has missed at least three attestations.\n\n%s", colorRed, colorReset)

		if !cliutils.Confirm("Have you removed the key from your own Validator Client and restarted it so that it is no longer active?") {
			fmt.Println("Cancelled.")
			return nil
		}
	} else {
		fmt.Println("NOTE: You have not requested to import the validator's private key into the Validator Client managed by the Smartnode. You will still be responsible for running and maintaining your own Validator Client with the validator's private key loaded, just as you are today.\n")
	}

	// Check deposit can be made
	canDeposit, err := rp.CanCreateVacantMinipool(amountWei, minNodeFee, salt, pubkey, importKey)
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
			fmt.Println("Node deposits are currently disabled.")
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
		fmt.Printf("%s**WARNING**: Can't verify the sync status of your consensus client.\nYOU WILL LOSE ETH if your minipool is activated before it is fully synced.\n"+
			"Reason: %s\n%s", colorRed, err, colorReset)
	} else {
		if syncResponse.BcStatus.PrimaryClientStatus.IsSynced {
			fmt.Printf("Your consensus client is synced, you may safely create a minipool.\n")
		} else if syncResponse.BcStatus.FallbackEnabled {
			if syncResponse.BcStatus.FallbackClientStatus.IsSynced {
				fmt.Printf("Your fallback consensus client is synced, you may safely create a minipool.\n")
			} else {
				fmt.Printf("%s**WARNING**: neither your primary nor fallback consensus clients are fully synced.\nYOU WILL LOSE ETH if your minipool is activated before they are fully synced.\n%s", colorRed, colorReset)
			}
		} else {
			fmt.Printf("%s**WARNING**: your primary consensus client is either not fully synced or offline and you do not have a fallback client configured.\nYOU WILL LOSE ETH if your minipool is activated before it is fully synced.\n%s", colorRed, colorReset)
		}
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canDeposit.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf(
		"You are about to deposit %.6f ETH to create a minipool with a minimum possible commission rate of %f%%.\n"+
			"%sARE YOU SURE YOU WANT TO DO THIS? Running a minipool is a long-term commitment, and this action cannot be undone!%s",
		math.RoundDown(eth.WeiToEth(amountWei), 6),
		minNodeFee*100,
		colorYellow,
		colorReset))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Make deposit
	response, err := rp.CreateVacantMinipool(amountWei, minNodeFee, salt, pubkey, importKey)
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

	if importKey {
		fmt.Print("Restarting Validator Client...")
		// Restart the VC
		_, err := rp.RestartVc()
		if err != nil {
			fmt.Printf("failed!\n%sWARNING: error restarting validator client: %s\n\nPlease restart it manually so it picks up the new validator key for your minipool.%s", colorYellow, err.Error(), colorReset)
		} else {
			fmt.Println(" done!\n")
		}
	}

	// Log & return
	fmt.Println("Your minipool was made successfully!")
	fmt.Printf("Your new minipool's address is: %s\n", response.MinipoolAddress)

	fmt.Printf("You can now upgrade your validator's withdrawal credentials to the following:\n\n\t%s\n\n", response.WithdrawalCredentials)
	fmt.Printf("It has entered the scrub check, where it will hold for %s.", response.ScrubPeriod)
	fmt.Println("You can watch its progress using `rocketpool service logs node`.")
	fmt.Println("Once the scrub check period has passed, your node will automatically promote it to an active minipool.")

	return nil

}
