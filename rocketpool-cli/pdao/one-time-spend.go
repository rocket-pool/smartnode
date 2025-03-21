package pdao

import (
	"fmt"
	"math/big"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/urfave/cli"
)

func proposeOneTimeSpend(c *cli.Context) error {
	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check for the raw flag
	rawEnabled := c.Bool("raw")

	// Get the invoice ID
	invoiceID := c.String("invoice-id")
	if invoiceID == "" {
		invoiceID = cliutils.Prompt("Please enter an invoice ID for this spend: ", "^\\s*\\S+\\s*$", "Invalid ID")
	}

	// Get the recipient
	recipientString := c.String("recipient")
	if recipientString == "" {
		recipientString = cliutils.Prompt("Please enter a recipient address for this spend:", "^0x[0-9a-fA-F]{40}$", "Invalid recipient address")
	}
	recipient, err := cliutils.ValidateAddress("recipient", recipientString)
	if err != nil {
		return err
	}

	// Get the amount string
	amountString := c.String("amount")
	if amountString == "" {
		if rawEnabled {
			amountString = cliutils.Prompt(fmt.Sprintf("Please enter an amount of RPL to send to %s as a wei amount:", recipientString), "^[0-9]+$", "Invalid amount")
		} else {
			amountString = cliutils.Prompt(fmt.Sprintf("Please enter an amount of RPL to send to %s:", recipientString), "^[0-9]+(\\.[0-9]+)?$", "Invalid amount")
		}
	}

	// Parse the amount
	var amount *big.Int
	if rawEnabled {
		amount, err = cliutils.ValidateBigInt("amount", amountString)
	} else {
		amount, err = parseFloat(c, "amount", amountString, false)
	}
	if err != nil {
		return err
	}

	// Check submissions
	canResponse, err := rp.PDAOCanProposeOneTimeSpend(invoiceID, recipient, amount)
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
	err = gas.AssignMaxFeeAndLimit(canResponse.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm("Are you sure you want to propose this one-time spend of the Protocol DAO treasury?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Submit
	response, err := rp.PDAOProposeOneTimeSpend(invoiceID, recipient, amount, canResponse.BlockNumber)
	if err != nil {
		return err
	}

	fmt.Printf("Proposing one-time spend...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Println("Proposal successfully created.")
	return nil

}
