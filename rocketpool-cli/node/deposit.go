package node

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"

	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

// Config
const DefaultMaxNodeFeeSlippage = 0.01 // 1% below current network fee

func nodeDeposit(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

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
		selected, _ := cliutils.Select("Please choose an amount of ETH to deposit:", amountOptions)
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

	// Check deposit can be made
	canDeposit, err := rp.CanNodeDeposit(amountWei, minNodeFee, salt)
	if err != nil {
		return err
	}
	if !canDeposit.CanDeposit {
		fmt.Println("Cannot make node deposit:")
		if canDeposit.InsufficientBalanceWithoutCredit {
			nodeBalance := eth.WeiToEth(canDeposit.NodeBalance)
			fmt.Printf("There is not enough ETH in the staking pool to use your credit balance (it needs at least 1 ETH but only has %.2f ETH) and you don't have enough ETH in your wallet (%.6f ETH) to cover the deposit amount yourself. If you want to continue creating a minipool, you will either need to wait for the staking pool to have more ETH deposited or add more ETH to your node wallet.", eth.WeiToEth(canDeposit.DepositBalance), nodeBalance)
		}
		if canDeposit.InsufficientBalance {
			nodeBalance := eth.WeiToEth(canDeposit.NodeBalance)
			creditBalance := eth.WeiToEth(canDeposit.CreditBalance)
			fmt.Printf("The node's balance of %.6f ETH and credit balance of %.6f ETH are not enough to create a minipool with a %.1f ETH bond.", nodeBalance, creditBalance, amount)
		}
		if canDeposit.InsufficientRplStake {
			fmt.Printf("The node has not staked enough RPL to collateralize a new minipool with a bond of %d ETH (this also includes the RPL required to support any pending bond reductions).\n", int(amount))
		}
		if canDeposit.InvalidAmount {
			fmt.Println("The deposit amount is invalid.")
		}
		if canDeposit.UnbondedMinipoolsAtMax {
			fmt.Println("The node cannot create any more unbonded minipools.")
		}
		if canDeposit.DepositDisabled {
			fmt.Println("Node deposits are currently disabled.")
		}
		return nil
	}

	useCreditBalance := false
	fmt.Printf("You currently have %.2f ETH in your credit balance plus ETH staked on your behalf.\n", eth.WeiToEth(canDeposit.CreditBalance))
	if canDeposit.CreditBalance.Cmp(big.NewInt(0)) > 0 {
		if canDeposit.CanUseCredit {
			useCreditBalance = true
			// Get how much credit to use
			remainingAmount := big.NewInt(0).Sub(amountWei, canDeposit.CreditBalance)
			if remainingAmount.Cmp(big.NewInt(0)) > 0 {
				fmt.Printf("This deposit will use all %.6f ETH from your credit balance plus ETH staked on your behalf and %.6f ETH from your node.\n\n", eth.WeiToEth(canDeposit.CreditBalance), eth.WeiToEth(remainingAmount))
			} else {
				fmt.Printf("This deposit will use %.6f ETH from your credit balance plus ETH staked on your behalf and will not require any ETH from your node.\n\n", amount)
			}
		} else {
			fmt.Printf("%sNOTE: Your credit balance *cannot* currently be used to create a new minipool; there is not enough ETH in the staking pool to cover the initial deposit on your behalf (it needs at least 1 ETH but only has %.2f ETH).%s\nIf you want to continue creating this minipool now, you will have to pay for the full bond amount.\n\n", colorYellow, eth.WeiToEth(canDeposit.DepositBalance), colorReset)
		}
	}
	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm("Would you like to continue?")) {
		fmt.Println("Cancelled.")
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
			"%sARE YOU SURE YOU WANT TO DO THIS? Exiting this minipool and retrieving your capital cannot be done until your minipool has been *active* on the Beacon Chain for 256 epochs (approx. 27 hours).%s\n",
		math.RoundDown(eth.WeiToEth(amountWei), 6),
		minNodeFee*100,
		colorYellow,
		colorReset))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Make deposit
	response, err := rp.NodeDeposit(amountWei, minNodeFee, salt, useCreditBalance, true)
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

	// Log & return
	fmt.Printf("The node deposit of %.6f ETH was made successfully!\n", math.RoundDown(eth.WeiToEth(amountWei), 6))
	fmt.Printf("Your new minipool's address is: %s\n", response.MinipoolAddress)
	fmt.Printf("The validator pubkey is: %s\n\n", response.ValidatorPubkey.Hex())

	fmt.Println("Your minipool is now in Initialized status.")
	fmt.Println("Once the remaining ETH has been assigned to your minipool from the staking pool, it will move to Prelaunch status.")
	fmt.Printf("After that, it will move to Staking status once %s have passed.\n", response.ScrubPeriod)
	fmt.Println("You can watch its progress using `rocketpool service logs node`.")

	return nil

}
