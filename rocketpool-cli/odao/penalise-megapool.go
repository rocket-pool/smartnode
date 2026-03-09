package odao

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

func penaliseMegapool(megapoolAddress common.Address, block *big.Int, yes bool) error {
	// Get RP client
	rp, err := rocketpool.NewClient().WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get amount to repay
	amountStr := prompt.Prompt("Enter the amount to penalise the megapool (in ETH):", "^\\d+(\\.\\d+)?$", "Invalid amount")

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return fmt.Errorf("Invalid amount '%s': %w\n", amountStr, err)
	}

	amountWei := eth.EthToWei(amount)
	// Check megapool debt can be repaid
	canPenalise, err := rp.CanPenaliseMegapool(megapoolAddress, block, amountWei)
	if err != nil {
		return err
	}

	if !canPenalise.CanPenalise {
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canPenalise.GasInfo, rp, yes)
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(yes || prompt.Confirm("Are you sure you want to penalise %.6f megapool %s at block %s?", math.RoundDown(eth.WeiToEth(amountWei), 6), megapoolAddress, block)) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Penalise the megapool
	response, err := rp.PenaliseMegapool(megapoolAddress, block, amountWei)
	if err != nil {
		return err
	}

	fmt.Printf("Penalising megapool...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully penalised megapool %s with %.6f debt.\n", megapoolAddress, math.RoundDown(eth.WeiToEth(amountWei), 6))
	return nil

}
