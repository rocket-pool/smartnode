package megapool

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

// Config
const (
	colorReset  string = "\033[0m"
	colorRed    string = "\033[31m"
	colorGreen  string = "\033[32m"
	colorYellow string = "\033[33m"
	maxCount    uint64 = 35
)

func nodeMegapoolDeposit(c *cli.Context) error {

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

	fmt.Println("Your eth2 client is on the correct network.")
	fmt.Println()

	saturnDeployed, err := rp.IsSaturnDeployed()
	if err != nil {
		return err
	}

	if !saturnDeployed.IsSaturnDeployed {
		fmt.Println("This command is only available after Saturn 1 is deployed.")
		return nil
	}

	useExpressTicket := false

	var wg errgroup.Group
	var expressTicketCount uint64
	var queueDetails api.GetQueueDetailsResponse
	var amount float64
	// Get the express ticket count
	wg.Go(func() error {
		expressTicket, err := rp.GetExpressTicketCount()
		if err != nil {
			return err
		}
		expressTicketCount = expressTicket.Count
		return nil
	})
	wg.Go(func() error {
		queueDetails, err = rp.GetQueueDetails()
		if err != nil {
			return err
		}
		return nil
	})
	wg.Go(func() error {
		settings, err := rp.PDAOGetSettings()
		if err != nil {
			return err
		}
		amount = settings.Node.ReducedBond
		return nil
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return err
	}

	count := c.Uint64("count")

	// If the count was not provided, prompt the user for the number of deposits
	for count == 0 || count > maxCount {
		countStr := prompt.Prompt(fmt.Sprintf("How many validators would you like to create? (max: %d)", maxCount), "^\\d+$", "Invalid number.")
		count, err = strconv.ParseUint(countStr, 10, 64)
		if err != nil {
			fmt.Println("Invalid number. Please try again.")
			continue
		}
		break
	}

	if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("%sNOTE: You are about to create %d new megapool validators, each with a %.0f ETH deposit (total: %.0f ETH).%s\nWould you like to continue?", colorYellow, count, amount, amount*float64(count), colorReset))) {
		fmt.Println("Cancelled.")
		return nil
	}

	fmt.Printf("There are %d validator(s) on the express queue.\n", queueDetails.ExpressLength)
	fmt.Printf("There are %d validator(s) on the standard queue.\n", queueDetails.StandardLength)
	fmt.Printf("The express queue rate is %d.\n\n", queueDetails.ExpressRate)

	if c.Bool("use-express-ticket") {
		if expressTicketCount > 0 {
			useExpressTicket = true
		} else {
			fmt.Println("You do not have any express tickets available.")
			return nil
		}
	} else {
		if expressTicketCount > 0 {
			fmt.Printf("You have %d express tickets available.", expressTicketCount)
			fmt.Println()
			// Prompt for confirmation
			if c.Bool("yes") || prompt.Confirm("Would you like to use your express tickets?") {
				useExpressTicket = true
			}
		}
	}

	amountWei := eth.EthToWei(amount)
	minNodeFee := 0.0

	// Check deposit can be made
	var canDeposit api.CanNodeDepositResponse
	if count > 1 {
		canDeposit, err = rp.CanNodeDeposits(count, amountWei, minNodeFee, big.NewInt(0), useExpressTicket)
	} else {
		canDeposit, err = rp.CanNodeDeposit(amountWei, minNodeFee, big.NewInt(0), useExpressTicket)
	}
	if err != nil {
		return err
	}
	if !canDeposit.CanDeposit {
		fmt.Printf("Cannot make %d node deposits:\n", count)

		if canDeposit.InsufficientBalanceWithoutCredit {
			nodeBalance := eth.WeiToEth(canDeposit.NodeBalance)
			fmt.Printf("There is not enough ETH in the staking pool to use your credit balance (it needs at least 1 ETH but only has %.2f ETH) and you don't have enough ETH in your wallet (%.6f ETH) to cover the deposit amount yourself. If you want to continue creating a minipool, you will either need to wait for the staking pool to have more ETH deposited or add more ETH to your node wallet.", eth.WeiToEth(canDeposit.DepositBalance), nodeBalance)
		}
		if canDeposit.InsufficientBalance {
			nodeBalance := eth.WeiToEth(canDeposit.NodeBalance)
			creditBalance := eth.WeiToEth(canDeposit.CreditBalance)
			if count > 1 {
				fmt.Printf("The node's balance of %.6f ETH and credit balance of %.6f ETH are not enough to create %d megapool validators with a %.1f ETH bond each (total: %.1f ETH).", nodeBalance, creditBalance, count, amount, amount*float64(count))
			} else {
				fmt.Printf("The node's balance of %.6f ETH and credit balance of %.6f ETH are not enough to create a megapool validator with a %.1f ETH bond.", nodeBalance, creditBalance, amount)
			}
		}
		if canDeposit.InvalidAmount {
			fmt.Println("The deposit amount is invalid.")
		}
		if canDeposit.DepositDisabled {
			fmt.Println("Node deposits are currently disabled.")
		}
		return nil
	}

	useCreditBalance := false
	totalAmountWei := big.NewInt(0).Mul(amountWei, big.NewInt(int64(count)))
	fmt.Printf("You currently have %.2f ETH in your credit balance plus ETH staked on your behalf.\n", eth.WeiToEth(canDeposit.CreditBalance))
	if canDeposit.CreditBalance.Cmp(big.NewInt(0)) > 0 {
		if canDeposit.CanUseCredit {
			useCreditBalance = true
			// Get how much credit to use
			remainingAmount := big.NewInt(0).Sub(totalAmountWei, canDeposit.CreditBalance)
			if remainingAmount.Cmp(big.NewInt(0)) > 0 {
				fmt.Printf("This deposit will use all %.6f ETH from your credit balance plus ETH staked on your behalf and %.6f ETH from your node.\n\n", eth.WeiToEth(canDeposit.CreditBalance), eth.WeiToEth(remainingAmount))
			} else {
				fmt.Printf("This deposit will use %.6f ETH from your credit balance plus ETH staked on your behalf and will not require any ETH from your node.\n\n", amount)
			}
		} else {
			fmt.Printf("%sNOTE: Your credit balance *cannot* currently be used to create a new megapool validator; there is not enough ETH in the staking pool to cover the initial deposit on your behalf (it needs at least 1 ETH but only has %.2f ETH).%s\nIf you want to continue creating this megapool validator now, you will have to pay for the full bond amount.\n\n", colorYellow, eth.WeiToEth(canDeposit.DepositBalance), colorReset)
		}
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm("Would you like to continue?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Check to see if eth2 is synced
	colorReset := "\033[0m"
	colorRed := "\033[31m"
	colorYellow := "\033[33m"
	syncResponse, err := rp.NodeSync()
	if err != nil {
		fmt.Printf("%s**WARNING**: Can't verify the sync status of your consensus client.\nYOU WILL LOSE ETH if your megapool validator is activated before it is fully synced.\n"+
			"Reason: %s\n%s", colorRed, err, colorReset)
	} else {
		if syncResponse.BcStatus.PrimaryClientStatus.IsSynced {
			fmt.Printf("Your consensus client is synced, you may safely create a megapool validator.\n")
		} else if syncResponse.BcStatus.FallbackEnabled {
			if syncResponse.BcStatus.FallbackClientStatus.IsSynced {
				fmt.Printf("Your fallback consensus client is synced, you may safely create a megapool validator.\n")
			} else {
				fmt.Printf("%s**WARNING**: neither your primary nor fallback consensus clients are fully synced.\nYOU WILL LOSE ETH if your megapool validator is activated before they are fully synced.\n%s", colorRed, colorReset)
			}
		} else {
			fmt.Printf("%s**WARNING**: your primary consensus client is either not fully synced or offline and you do not have a fallback client configured.\nYOU WILL LOSE ETH if your megapool validator is activated before it is fully synced.\n%s", colorRed, colorReset)
		}
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canDeposit.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation

	if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf(
		"You are about to deposit %.6f ETH to create %d new megapool validators (%.6f ETH total).\n"+
			"%sARE YOU SURE YOU WANT TO DO THIS? Exiting these validators and retrieving your capital cannot be done until each validator has been *active* on the Beacon Chain for 256 epochs (approx. 27 hours).%s\n",
		math.RoundDown(eth.WeiToEth(amountWei), 6),
		count,
		math.RoundDown(eth.WeiToEth(amountWei), 6)*float64(count),
		colorYellow,
		colorReset))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Make deposit(s)
	if count == 1 {
		// Single deposit
		response, err := rp.NodeDeposit(amountWei, minNodeFee, big.NewInt(0), useCreditBalance, useExpressTicket, true)
		if err != nil {
			return err
		}

		// Log and wait for the megapool validator deposit
		fmt.Printf("Creating megapool validator...\n")
		cliutils.PrintTransactionHash(rp, response.TxHash)
		_, err = rp.WaitForTransaction(response.TxHash)
		if err != nil {
			return err
		}

		// Log & return
		fmt.Printf("The node deposit of %.6f ETH was made successfully!\n", math.RoundDown(eth.WeiToEth(amountWei), 6))
		fmt.Printf("The validator pubkey is: %s\n\n", response.ValidatorPubkey.Hex())

		fmt.Println("The new megapool validator has been created.")
		fmt.Println("Once your validator progresses through the queue, ETH will be assigned and a 1 ETH prestake submitted.")
		fmt.Printf("After the prestake, your node will automatically perform a stake transaction, to complete the progress.")
		fmt.Println("")
		fmt.Println("To check the status of your validators use `rocketpool megapool validators`")
		fmt.Println("To monitor the stake transaction use `rocketpool service logs node`")
	} else {
		// Multiple deposits
		responses, err := rp.NodeDeposits(count, amountWei, minNodeFee, big.NewInt(0), useCreditBalance, useExpressTicket, true)
		if err != nil {
			return err
		}

		// Log and wait for the megapool validator deposits
		fmt.Printf("Creating %d megapool validators in a single transaction...\n", count)
		cliutils.PrintTransactionHash(rp, responses.TxHash)
		_, err = rp.WaitForTransaction(responses.TxHash)
		if err != nil {
			return err
		}

		// Log & return
		fmt.Printf("The node deposits of %.6f ETH each (%.6f ETH total) were made successfully!\n",
			math.RoundDown(eth.WeiToEth(amountWei), 6),
			math.RoundDown(eth.WeiToEth(amountWei), 6)*float64(count))
		fmt.Printf("Validator pubkeys:\n")
		for i, pubkey := range responses.ValidatorPubkeys {
			fmt.Printf("  %d. %s\n", i+1, pubkey.Hex())
		}
		fmt.Println()

		fmt.Printf("The %d new megapool validators have been created.\n", count)
		fmt.Println("Once your validators progress through the queue, ETH will be assigned and a 1 ETH prestake submitted for each.")
		fmt.Printf("After the prestake, your node will automatically perform a stake transaction for each validator, to complete the progress.")
		fmt.Println("")
		fmt.Println("To check the status of your validators use `rocketpool megapool validators`")
		fmt.Println("To monitor the stake transactions use `rocketpool service logs node`")
	}

	return nil

}
