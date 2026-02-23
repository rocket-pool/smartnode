package queue

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

func assignDeposits(c *cli.Context) error {
	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get queue details
	queueDetails, err := rp.GetQueueDetails()
	if err != nil {
		return err
	}

	queueStatus, err := rp.QueueStatus()
	if err != nil {
		return err
	}

	validatorDeposit := eth.EthToWei(32)
	if queueDetails.TotalLength == 0 {
		fmt.Println("There are no validators waiting in the queue.")
		return nil
	}

	// Calculate how many validator assignments are possible given the deposit pool balance
	// and the deposit required per validator
	depositPoolBalance := queueStatus.DepositPoolBalance
	assignmentsPossible := new(big.Int).Div(depositPoolBalance, validatorDeposit).Uint64()

	// The effective max is the lesser of what the deposit pool can fund and what's in the queue
	maxAssignable := min(assignmentsPossible, uint64(queueDetails.TotalLength))

	fmt.Printf("There are %d validator(s) in the queue (%d express, %d standard).\n",
		queueDetails.TotalLength, queueDetails.ExpressLength, queueDetails.StandardLength)
	fmt.Printf("The deposit pool can fund up to %d assignment(s).\n", assignmentsPossible)

	if maxAssignable == 0 {
		fmt.Println("The deposit pool balance is insufficient to assign any deposits.")
		return nil
	}

	// Prompt for max validators
	var maxValidators uint64
	for {
		maxValidatorsStr := prompt.Prompt(
			fmt.Sprintf("How many deposits would you like to assign? (max: %d)", maxAssignable),
			"^\\d+$", "Invalid number.")
		maxValidators, err = strconv.ParseUint(maxValidatorsStr, 0, 64)
		if err != nil {
			fmt.Println("Invalid number. Please try again.")
			continue
		}
		if maxValidators == 0 || maxValidators > maxAssignable {
			fmt.Printf("Please enter a number between 1 and %d.\n", maxAssignable)
			continue
		}
		break
	}

	// Check deposits can be assigned
	canAssign, err := rp.CanAssignDeposits(uint32(maxValidators))
	if err != nil {
		return err
	}
	if !canAssign.CanAssign {
		fmt.Println("Cannot assign deposits:")
		if canAssign.AssignDepositsDisabled {
			fmt.Println("Deposit assignments are currently disabled.")
		}
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canAssign.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to assign %d validators?", maxValidators))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Assign deposits
	response, err := rp.AssignDeposits(uint32(maxValidators))
	if err != nil {
		return err
	}

	fmt.Printf("Assigning deposits...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Println("Deposits were successfully assigned.")
	return nil

}
