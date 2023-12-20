package tx

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/smartnode/rocketpool-cli/flags"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/client"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/gas"
	"github.com/urfave/cli"
)

func HandleTx(c *cli.Context, rp *client.Client, txInfo *core.TransactionInfo, confirmMessage string, submissionMessage string) error {
	// Print the TX data if requested
	if c.GlobalBool(flags.PrintTxDataFlag) {
		fmt.Println("TX Data:")
		fmt.Printf("\tTo:       %s\n", txInfo.To.Hex())
		fmt.Printf("\tData:     %s\n", hexutil.Encode(txInfo.Data))
		fmt.Printf("\tValue:    %s\n", txInfo.Value.String())
		fmt.Printf("\tEst. Gas: %d\n", txInfo.GasInfo.EstGasLimit)
		fmt.Printf("\tSafe Gas: %d\n", txInfo.GasInfo.SafeGasLimit)
		return nil
	}

	// Assign max fees
	maxFee, maxPrioFee, err := gas.GetMaxFees(c, rp, txInfo.GasInfo)
	if err != nil {
		return fmt.Errorf("error getting fee information: %w", err)
	}

	// Check the nonce flag
	var nonce *big.Int
	if c.GlobalIsSet(flags.NonceFlag) {
		nonce = big.NewInt(0).SetUint64(c.GlobalUint64(flags.NonceFlag))
	}

	// Create the submission from the TX info
	submission, _ := core.CreateTxSubmissionFromInfo(txInfo, nil)

	// Sign only (no submission) if requested
	if c.GlobalBool(flags.SignTxOnlyFlag) {
		response, err := rp.Api.Tx.SignTx(submission, nonce, maxFee, maxPrioFee)
		if err != nil {
			return fmt.Errorf("error signing transaction: %w", err)
		}
		fmt.Println("Signed transaction:")
		fmt.Println(response.Data.SignedTx)
		return nil
	}

	// Confirm submission
	if !(c.Bool(flags.YesFlag) || utils.Confirm(confirmMessage)) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Submit it
	fmt.Println(submissionMessage)
	response, err := rp.Api.Tx.SubmitTx(submission, nonce, maxFee, maxPrioFee)
	if err != nil {
		return fmt.Errorf("error submitting transaction: %w", err)
	}

	// Wait for it
	utils.PrintTransactionHash(rp, response.Data.TxHash)
	if _, err = rp.Api.Tx.WaitForTransaction(response.Data.TxHash); err != nil {
		return fmt.Errorf("error waiting for transaction: %w", err)
	}
	return nil
}
