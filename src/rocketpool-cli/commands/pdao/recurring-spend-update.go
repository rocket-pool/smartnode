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

func proposeRecurringSpendUpdate(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Check for the raw flag
	rawEnabled := c.Bool(utils.RawFlag.Name)

	// Get the contract name
	contractName := c.String(contractNameFlag.Name)
	if contractName == "" {
		contractName = utils.Prompt("Please enter a contract name for this recurring payment: (no spaces)", "^\\S+$", "Invalid ID")
	}

	// Get the recipient
	recipientString := c.String(recipientFlag.Name)
	if recipientString == "" {
		recipientString = utils.Prompt("Please enter a recipient address for this recurring payment:", "^0x[0-9a-fA-F]{40}$", "Invalid recipient address")
	}
	recipient, err := input.ValidateAddress("recipient", recipientString)
	if err != nil {
		return err
	}

	// Get the amount string
	var amount *big.Int
	amountString := c.String(amountPerPeriodFlag.Name)
	if amountString == "" {
		if rawEnabled {
			amountString = utils.Prompt(fmt.Sprintf("Please enter an amount of RPL to send to %s per period as a wei amount:", recipientString), "^[0-9]+$", "Invalid amount")
		} else {
			amountString = utils.Prompt(fmt.Sprintf("Please enter an amount of RPL to send to %s per period:", recipientString), "^[0-9]+(\\.[0-9]+)?$", "Invalid amount")
		}
	}
	if rawEnabled {
		amount, err = input.ValidateBigInt("amount-per-period", amountString)
	} else {
		amount, err = utils.ParseFloat(c, "amount-per-period", amountString, false)
	}
	if err != nil {
		return err
	}

	// Get the period length
	periodLengthString := c.String(periodLengthFlag.Name)
	if periodLengthString == "" {
		periodLengthString = utils.Prompt("Please enter the length of each payment period in hours / minutes / seconds (e.g., 168h0m0s):", "^.+$", "Invalid period length")
	}
	periodLength, err := input.ValidateDuration("period-length", periodLengthString)
	if err != nil {
		return err
	}

	// Get the number of periods
	numPeriods := c.Uint64(numberOfPeriodsFlag.Name)
	if !c.IsSet(numberOfPeriodsFlag.Name) {
		numPeriodsString := utils.Prompt("Please enter the total number of payment periods:", "^[0-9]+$", "Invalid number of periods")
		numPeriods, err = input.ValidateUint("number-of-periods", numPeriodsString)
		if err != nil {
			return err
		}
	}

	// Build the TX
	response, err := rp.Api.PDao.RecurringSpendUpdate(contractName, recipient, amount, periodLength, numPeriods)
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
		"Are you sure you want to propose updating this recurring spend of the Protocol DAO treasury?",
		"recurring spend update proposal",
		"Proposing recurring spend update...",
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
