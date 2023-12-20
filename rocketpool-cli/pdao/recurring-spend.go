package pdao

import (
	"fmt"
	"math/big"
	"time"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/urfave/cli"
)

func proposeRecurringSpend(c *cli.Context) error {
	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check for Houston
	houston, err := rp.IsHoustonDeployed()
	if err != nil {
		return fmt.Errorf("error checking if Houston has been deployed: %w", err)
	}
	if !houston.IsHoustonDeployed {
		fmt.Println("This command cannot be used until Houston has been deployed.")
		return nil
	}

	// Check for the raw flag
	rawEnabled := c.Bool("raw")

	// Get the contract name
	contractName := c.String("contract-name")
	if contractName == "" {
		contractName = cliutils.Prompt("Please enter a contract name for this recurring payment:", "^/S+$", "Invalid ID")
	}

	// Get the recipient
	recipientString := c.String("recipient")
	if recipientString == "" {
		recipientString = cliutils.Prompt("Please enter a recipient address for this recurring payment:", "^0x[0-9a-fA-F]{40}$", "Invalid recipient address")
	}
	recipient, err := cliutils.ValidateAddress("recipient", recipientString)
	if err != nil {
		return err
	}

	// Get the amount string
	amountString := c.String("amount-per-period")
	if amountString == "" {
		if rawEnabled {
			amountString = cliutils.Prompt("Please enter an amount of RPL to send to %s per period as a wei amount:", "^[0-9]+$", "Invalid amount")
		} else {
			amountString = cliutils.Prompt("Please enter an amount of RPL to send to %s per period:", "^[0-9]+(\\.[0-9]+)?$", "Invalid amount")
		}
	}

	// Parse the amount
	var amount *big.Int
	if rawEnabled {
		amount, err = cliutils.ValidateBigInt("amount-per-period", amountString)
	} else {
		amount, err = parseFloat(c, "amount-per-period", amountString)
	}
	if err != nil {
		return err
	}

	// Get the start time
	startTimeString := c.String("start-time")
	if startTimeString == "" {
		startTimeString = cliutils.Prompt("Please enter the time that the recurring payment will start (as a UNIX timestamp):", "^[0-9]+$", "Invalid start time")
	}
	startTimeUnix, err := cliutils.ValidateUint("start-time", startTimeString)
	if err != nil {
		return err
	}
	startTime := time.Unix(int64(startTimeUnix), 0)
	if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("The provided timestamp corresponds to %s - is this correct?", startTime.UTC().String()))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Get the period length
	periodLengthString := c.String("period-length")
	if periodLengthString == "" {
		periodLengthString = cliutils.Prompt("Please enter the length of each payment period in hours / minutes / seconds (e.g., 168h0m0s):", "^+$", "Invalid period length")
	}
	periodLength, err := cliutils.ValidateDuration("period-length", periodLengthString)
	if err != nil {
		return err
	}

	// Get the number of periods
	numPeriodsString := c.String("number-of-periods")
	if numPeriodsString == "" {
		numPeriodsString = cliutils.Prompt("Please enter the total number of payment periods:", "^[0-9]+$", "Invalid number of periods")
	}
	numPeriods, err := cliutils.ValidateUint("number-of-periods", numPeriodsString)
	if err != nil {
		return err
	}

	// Check submissions
	canResponse, err := rp.PDAOCanProposeRecurringSpend(contractName, recipient, amount, periodLength, startTime, numPeriods)
	if err != nil {
		return err
	}

	// Assign max fee
	err = gas.AssignMaxFeeAndLimit(canResponse.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm("Are you sure you want to propose this recurring spend of the Protocol DAO treasury?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Submit
	response, err := rp.PDAOProposeRecurringSpend(contractName, recipient, amount, periodLength, startTime, numPeriods, canResponse.BlockNumber)
	if err != nil {
		return err
	}

	fmt.Printf("Proposing recurring spend...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Println("Proposal successfully created.")
	return nil

}
