package pdao

import (
	"fmt"
	"math/big"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

func proposeOneTimeSpend(invoiceIDFlag string, recipientFlag string, amountFlag string, customMessageFlag string, rawEnabled bool, yes bool) error {
	// Get RP client
	rp, err := rocketpool.NewClient().WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get the invoice ID
	if invoiceIDFlag == "" {
		invoiceIDFlag = prompt.Prompt("Please enter an invoice ID for this spend: ", "^\\s*\\S+\\s*$", "Invalid ID")
	}

	// Get the recipient
	if recipientFlag == "" {
		recipientFlag = prompt.Prompt("Please enter a recipient address for this spend:", "^0x[0-9a-fA-F]{40}$", "Invalid recipient address")
	}
	recipient, err := cliutils.ValidateAddress("recipient", recipientFlag)
	if err != nil {
		return err
	}

	// Get the amount string
	if amountFlag == "" {
		if rawEnabled {
			amountFlag = prompt.Prompt(fmt.Sprintf("Please enter an amount of RPL to send to %s as a wei amount:", recipientFlag), "^[0-9]+$", "Invalid amount")
		} else {
			amountFlag = prompt.Prompt(fmt.Sprintf("Please enter an amount of RPL to send to %s:", recipientFlag), "^[0-9]+(\\.[0-9]+)?$", "Invalid amount")
		}
	}

	// Parse the amount
	var amount *big.Int
	if rawEnabled {
		amount, err = cliutils.ValidateBigInt("amount", amountFlag)
	} else {
		amount, err = cliutils.ValidateFloat(rawEnabled, "amount", amountFlag, false, yes)
	}
	if err != nil {
		return err
	}

	// Get the custom message
	var message string
	if customMessageFlag == "" {
		message = prompt.Prompt("Please enter a custom message for this one-time spend (no blank spaces):", "^\\S*$", "Invalid message")
	}
	if message == "" {
		message = fmt.Sprintf("one-time-spend-for-invoice-%s", invoiceIDFlag)
	}

	// Check submissions
	canResponse, err := rp.PDAOCanProposeOneTimeSpend(invoiceIDFlag, recipient, amount, message)
	if err != nil {
		return err
	}
	if !canResponse.CanPropose {
		fmt.Println("Cannot propose one time spend:")
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
	if prompt.Declined(yes, "Are you sure you want to propose this one-time spend of the Protocol DAO treasury?") {
		fmt.Println("Cancelled.")
		return nil
	}

	// Submit
	response, err := rp.PDAOProposeOneTimeSpend(invoiceIDFlag, recipient, amount, canResponse.BlockNumber, message)
	if err != nil {
		return err
	}

	fmt.Println("Proposing one-time spend...")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Println("Proposal successfully created.")
	return nil

}
