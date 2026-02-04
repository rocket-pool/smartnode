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

	var wg errgroup.Group
	var expressTicketCount uint64
	var queueDetails api.GetQueueDetailsResponse
	var status api.MegapoolStatusResponse
	totalBondRequirement := big.NewInt(0)
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

	// Get the megapool status
	wg.Go(func() error {
		status, err = rp.MegapoolStatus(false)
		if err != nil {
			return err
		}
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
	}

	bondedEth := status.Megapool.NodeBond
	if bondedEth == nil {
		bondedEth = big.NewInt(0)
	}
	queuedBondEth := status.Megapool.NodeQueuedBond
	if queuedBondEth == nil {
		queuedBondEth = big.NewInt(0)
	}
	bondedEth = bondedEth.Add(bondedEth, queuedBondEth)
	megapoolBondedEth := big.NewInt(0).Set(bondedEth)
	lastBondAdded := big.NewInt(0)
	// Iterate through the deposits and get the bond requirement for each
	for i := uint64(1); i <= count; i++ {
		bondedEth = bondedEth.Add(bondedEth, lastBondAdded)
		activeValidatorCount := status.Megapool.ActiveValidatorCount
		bondRequirementResponse, err := rp.GetBondRequirement(i + uint64(activeValidatorCount))
		if err != nil {
			return err
		}

		lastBondAdded = bondRequirementResponse.BondRequirement
		// Find the bond requirement for the next validator
		nextBondRequirement := bondRequirementResponse.BondRequirement.Sub(bondRequirementResponse.BondRequirement, bondedEth)
		if nextBondRequirement.Cmp(eth.EthToWei(1)) < 0 {
			nextBondRequirement = eth.EthToWei(1)
		} else if nextBondRequirement.Cmp(eth.EthToWei(32)) > 0 {
			nextBondRequirement = eth.EthToWei(32)
		}
		totalBondRequirement = totalBondRequirement.Add(totalBondRequirement, nextBondRequirement)
	}

	totalBondRequirementEth := eth.WeiToEth(totalBondRequirement)
	// Show the node bond and the total bond requirement
	fmt.Printf("The node is currently bonded with %.2f ETH.\n", eth.WeiToEth(megapoolBondedEth))
	fmt.Printf("The total bond requirement is %.2f ETH.\n", totalBondRequirementEth)
	fmt.Println()

	if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("%sNOTE: You are about to create %d new megapool validators, requiring a total of: %.2f ETH).%s\nWould you like to continue?", colorYellow, count, totalBondRequirementEth, colorReset))) {
		fmt.Println("Cancelled.")
		return nil
	}

	fmt.Printf("There are %d validator(s) on the express queue.\n", queueDetails.ExpressLength)
	fmt.Printf("There are %d validator(s) on the standard queue.\n", queueDetails.StandardLength)
	fmt.Printf("The express queue rate is %d.\n\n", queueDetails.ExpressRate)

	expressTickets := c.Int64("express-tickets")
	if expressTickets >= 0 {
		if expressTicketCount < uint64(expressTickets) {
			expressTickets = int64(expressTicketCount)
		}
	}
	if expressTicketCount == 0 {
		expressTickets = int64(0)
	}
	if expressTicketCount > 0 && expressTickets < 0 {
		// Prompt for the number of express tickets to use
		for expressTickets == -1 || uint64(expressTickets) > expressTicketCount {
			expressTicketsStr := prompt.Prompt(fmt.Sprintf("How many express tickets would you like to use? (max: %d)", expressTicketCount), "^\\d+$", "Invalid number.")
			expressTickets, err = strconv.ParseInt(expressTicketsStr, 10, 64)
			if err != nil {
				fmt.Println("Invalid number. Please try again.")
			}
		}
	}

	minNodeFee := 0.0

	// Check deposit can be made
	canDeposit, err := rp.CanNodeDeposits(count, totalBondRequirement, minNodeFee, big.NewInt(0), uint64(expressTickets))
	if err != nil {
		return err
	}
	if !canDeposit.CanDeposit {
		fmt.Printf("Cannot make %d node deposits:\n", count)
		if canDeposit.NodeHasDebt {
			fmt.Println("The node has debt. You must repay the debt before creating a new validator. Use the `rocketpool megapool repay-debt` command to repay the debt.")
		}
		if canDeposit.InsufficientBalanceWithoutCredit {
			nodeBalance := eth.WeiToEth(canDeposit.NodeBalance)
			fmt.Printf("There is not enough ETH in the staking pool to use your credit balance (it needs at least 1 ETH but only has %.2f ETH) and you don't have enough ETH in your wallet (%.6f ETH) to cover the deposit amount yourself. If you want to continue creating a minipool, you will either need to wait for the staking pool to have more ETH deposited or add more ETH to your node wallet.", eth.WeiToEth(canDeposit.DepositBalance), nodeBalance)
		}
		if canDeposit.InsufficientBalance {
			nodeBalance := eth.WeiToEth(canDeposit.NodeBalance)
			creditBalance := eth.WeiToEth(canDeposit.CreditBalance)

			fmt.Printf("The node's balance of %.6f ETH and credit balance of %.6f ETH are not enough to create %d megapool validators with a total %.1f ETH bond.", nodeBalance, creditBalance, count, totalBondRequirementEth)

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
	totalAmountWei := totalBondRequirement
	fmt.Printf("You currently have %.2f ETH in your credit balance plus ETH staked on your behalf.\n", eth.WeiToEth(canDeposit.CreditBalance))
	if canDeposit.CreditBalance.Cmp(big.NewInt(0)) > 0 {
		if canDeposit.CanUseCredit {
			useCreditBalance = true
			// Get how much credit to use
			remainingAmount := big.NewInt(0).Sub(totalAmountWei, canDeposit.CreditBalance)
			if remainingAmount.Cmp(big.NewInt(0)) > 0 {
				fmt.Printf("This deposit will use all %.6f ETH from your credit balance plus ETH staked on your behalf and %.6f ETH from your node.\n\n", eth.WeiToEth(canDeposit.CreditBalance), eth.WeiToEth(remainingAmount))
			} else {
				fmt.Printf("This deposit will use %.6f ETH from your credit balance plus ETH staked on your behalf and will not require any ETH from your node.\n\n", totalBondRequirementEth)
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
		"You are about to deposit %.6f ETH to create %d new megapool validators.\n"+
			"%sARE YOU SURE YOU WANT TO DO THIS? %s\n",
		math.RoundDown(eth.WeiToEth(totalBondRequirement), 6),
		count,
		colorYellow,
		colorReset))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Make deposit(s)

	response, err := rp.NodeDeposits(count, totalBondRequirement, minNodeFee, big.NewInt(0), useCreditBalance, uint64(expressTickets), true)
	if err != nil {
		return err
	}
	// Log and wait for the megapool validator deposits
	fmt.Printf("Creating %d megapool validators ...\n", count)
	cliutils.PrintTransactionHash(rp, response.TxHash)
	_, err = rp.WaitForTransaction(response.TxHash)
	if err != nil {
		return err
	}

	// Log & return
	fmt.Printf("The node deposit of %.6f ETH total was made successfully!\n",
		math.RoundDown(eth.WeiToEth(totalBondRequirement), 6))
	fmt.Printf("Validator pubkeys:\n")
	for i, pubkey := range response.ValidatorPubkeys {
		fmt.Printf("  %d. %s\n", i+1, pubkey.Hex())
	}
	fmt.Println()

	fmt.Printf("The %d new megapool validators have been created.\n", count)
	fmt.Println("Once your validators progress through the queue, ETH will be assigned and a 1 ETH prestake submitted for each.")
	fmt.Printf("After the prestake, your node will automatically perform a stake transaction for each validator, to complete the progress.")
	fmt.Println("")
	fmt.Println("To check the status of your validators use `rocketpool megapool validators`")
	fmt.Println("To monitor the stake transactions use `rocketpool service logs node`")

	return nil

}
