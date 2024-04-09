package pdao

import (
	"fmt"
	"math/big"

	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
	"github.com/urfave/cli/v2"
)

var oneTimeSpendInvoiceFlag *cli.StringFlag = &cli.StringFlag{
	Name:    "invoice-id",
	Aliases: []string{"i"},
	Usage:   "The invoice ID / number for this spend",
}

func proposeOneTimeSpend(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Check for the raw flag
	rawEnabled := c.Bool(utils.RawFlag.Name)

	// Get the invoice ID
	invoiceID := c.String(oneTimeSpendInvoiceFlag.Name)
	if invoiceID == "" {
		invoiceID = utils.Prompt("Please enter an invoice ID for this spend: (no spaces)", "^\\S+$", "Invalid ID")
	}

	// Get the recipient
	recipientString := c.String(recipientFlag.Name)
	if recipientString == "" {
		recipientString = utils.Prompt("Please enter a recipient address for this spend:", "^0x[0-9a-fA-F]{40}$", "Invalid recipient address")
	}
	recipient, err := input.ValidateAddress("recipient", recipientString)
	if err != nil {
		return err
	}

	// Get the amount string
	amountString := c.String(amountFlag.Name)
	if amountString == "" {
		if rawEnabled {
			amountString = utils.Prompt(fmt.Sprintf("Please enter an amount of RPL to send to %s as a wei amount:", recipientString), "^[0-9]+$", "Invalid amount")
		} else {
			amountString = utils.Prompt(fmt.Sprintf("Please enter an amount of RPL to send to %s:", recipientString), "^[0-9]+(\\.[0-9]+)?$", "Invalid amount")
		}
	}

	// Parse the amount
	var amount *big.Int
	if rawEnabled {
		amount, err = input.ValidateBigInt("amount", amountString)
	} else {
		amount, err = utils.ParseFloat(c, "amount", amountString, false)
	}
	if err != nil {
		return err
	}

	// Build the TX
	response, err := rp.Api.PDao.OneTimeSpend(invoiceID, recipient, amount)
	if err != nil {
		return err
	}

	// Verify
	if !response.Data.CanPropose {
		fmt.Println("You cannot currently submit this proposal:")
		if response.Data.InsufficientRpl {
			fmt.Printf("You do not have enough unlocked RPL (proposals require locking %.6f RPL, but you only have %.6f RPL staked and unlocked).", eth.WeiToEth(response.Data.ProposalBond), eth.WeiToEth(big.NewInt(0).Sub(response.Data.StakedRpl, response.Data.LockedRpl)))
		}
		return nil
	}

	// Run the TX
	validated, err := tx.HandleTx(c, rp, response.Data.TxInfo,
		"Are you sure you want to propose this one-time spend of the Protocol DAO treasury?",
		"one-time-spend proposal",
		"Proposing one-time spend...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Println("Proposal successfully created.")
	return nil
}
