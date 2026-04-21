package pdao

import (
	"fmt"
	"math/big"
	"time"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

func proposeRecurringSpend(rawEnabled bool, contractName string, recipientString string, amountString string, startTimeUnix uint64, periodLengthString string, numPeriods uint64, customMessage string, yes bool) error {
	// Get RP client
	rp, err := rocketpool.NewClient().WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get the contract name
	if contractName == "" {
		contractName = prompt.Prompt("Please enter a contract name for this recurring payment:", "^\\s*\\S+\\s*$", "Invalid ID")
	}

	// Get the recipient
	if recipientString == "" {
		recipientString = prompt.Prompt("Please enter a recipient address for this recurring payment:", "^0x[0-9a-fA-F]{40}$", "Invalid recipient address")
	}
	recipient, err := cliutils.ValidateAddress("recipient", recipientString)
	if err != nil {
		return err
	}

	// Get the amount string
	if amountString == "" {
		if rawEnabled {
			amountString = prompt.Prompt(fmt.Sprintf("Please enter an amount of RPL to send to %s per period as a wei amount:", recipientString), "^[0-9]+$", "Invalid amount")
		} else {
			amountString = prompt.Prompt(fmt.Sprintf("Please enter an amount of RPL to send to %s per period:", recipientString), "^[0-9]+(\\.[0-9]+)?$", "Invalid amount")
		}
	}

	// Parse the amount
	var amount *big.Int
	if rawEnabled {
		amount, err = cliutils.ValidateBigInt("amount-per-period", amountString)
	} else {
		amount, err = cliutils.ValidateFloat(rawEnabled, "amount-per-period", amountString, false, yes)
	}
	if err != nil {
		return err
	}

	// Get the start time
	if startTimeUnix == 0 {
		startTimeString := prompt.Prompt("Please enter the time that the recurring payment will start (as a UNIX timestamp):", "^[0-9]+$", "Invalid start time")
		startTimeUnix, err = cliutils.ValidateUint("start-time", startTimeString)
		if err != nil {
			return err
		}
	}
	startTime := time.Unix(int64(startTimeUnix), 0)
	if prompt.Declined(yes, "The provided timestamp corresponds to %s - is this correct?", startTime.UTC().String()) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Get the period length
	if periodLengthString == "" {
		periodLengthString = prompt.Prompt("Please enter the length of each payment period in hours / minutes / seconds (e.g., 168h0m0s):", "^.+$", "Invalid period length")
	}
	periodLength, err := cliutils.ValidateDuration("period-length", periodLengthString)
	if err != nil {
		return err
	}

	// Get the number of periods
	if numPeriods == 0 {
		numPeriodsString := prompt.Prompt("Please enter the total number of payment periods:", "^[0-9]+$", "Invalid number of periods")
		numPeriods, err = cliutils.ValidateUint("number-of-periods", numPeriodsString)
		if err != nil {
			return err
		}
	}

	// Get the custom message
	if customMessage == "" {
		// Prompt for a custom message without blank spaces
		customMessage = prompt.Prompt("Please enter an optional message for this recurring payment (no blank spaces):", "^\\S*$", "Invalid message")
	}
	if customMessage == "" {
		customMessage = "recurring-payment-to-" + contractName
	}

	// Check submissions
	canResponse, err := rp.PDAOCanProposeRecurringSpend(contractName, recipient, amount, periodLength, startTime, numPeriods, customMessage)
	if err != nil {
		return err
	}
	if !canResponse.CanPropose {
		fmt.Println("Cannot propose recurring spend contract:")
		if canResponse.IsRplLockingDisallowed {
			fmt.Println("Please enable RPL locking using the command 'rocketpool node allow-rpl-locking' to raise proposals.")
		}
		return nil
	}

	// Assign max fee
	err = gas.AssignMaxFeeAndLimit(canResponse.GasInfo, rp, yes)
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if prompt.Declined(yes, "Are you sure you want to propose this recurring spend of the Protocol DAO treasury?") {
		fmt.Println("Cancelled.")
		return nil
	}

	// Submit
	response, err := rp.PDAOProposeRecurringSpend(contractName, recipient, amount, periodLength, startTime, numPeriods, canResponse.BlockNumber, customMessage)
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
