package tx

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/smartnode/rocketpool-cli/flags"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/client"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/gas"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

// Handle a transaction, either printing its details, signing it, or submitting it and waiting for it to be included
func HandleTx(c *cli.Context, rp *client.Client, txInfo *core.TransactionInfo, confirmMessage string, submissionMessage string) error {
	// Print the TX data if requested
	if c.Bool(flags.PrintTxDataFlag) {
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
	if c.IsSet(flags.NonceFlag) {
		nonce = big.NewInt(0).SetUint64(c.Uint64(flags.NonceFlag))
	}

	// Create the submission from the TX info
	submission, _ := core.CreateTxSubmissionFromInfo(txInfo, nil)

	// Sign only (no submission) if requested
	if c.Bool(flags.SignTxOnlyFlag) {
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

// Handle a batch of transactions, either printing their details, signing them, or submitting them and waiting for them to be included
func HandleTxBatch(c *cli.Context, rp *client.Client, txInfos []*core.TransactionInfo, confirmMessage string, submissionMessage string) error {
	// Print the TX data if requested
	if c.Bool(flags.PrintTxDataFlag) {
		for i, info := range txInfos {
			fmt.Printf("Data for TX %d:\n", i)
			fmt.Printf("\tTo:       %s\n", info.To.Hex())
			fmt.Printf("\tData:     %s\n", hexutil.Encode(info.Data))
			fmt.Printf("\tValue:    %s\n", info.Value.String())
			fmt.Printf("\tEst. Gas: %d\n", info.GasInfo.EstGasLimit)
			fmt.Printf("\tSafe Gas: %d\n", info.GasInfo.SafeGasLimit)
			fmt.Println()
		}
		return nil
	}

	// Assign max fees
	var gasInfo core.GasInfo
	for _, info := range txInfos {
		gasInfo.EstGasLimit += info.GasInfo.EstGasLimit
		gasInfo.SafeGasLimit += info.GasInfo.SafeGasLimit
	}
	maxFee, maxPrioFee, err := gas.GetMaxFees(c, rp, gasInfo)
	if err != nil {
		return fmt.Errorf("error getting fee information: %w", err)
	}

	// Check the nonce flag
	var nonce *big.Int
	if c.IsSet(flags.NonceFlag) {
		nonce = big.NewInt(0).SetUint64(c.Uint64(flags.NonceFlag))
	}

	// Create the submissions from the TX infos
	submissions := make([]*core.TransactionSubmission, len(txInfos))
	for i, info := range txInfos {
		submission, _ := core.CreateTxSubmissionFromInfo(info, nil)
		submissions[i] = submission
	}

	// Sign only (no submission) if requested
	if c.Bool(flags.SignTxOnlyFlag) {
		response, err := rp.Api.Tx.SignTxBatch(submissions, nonce, maxFee, maxPrioFee)
		if err != nil {
			return fmt.Errorf("error signing transactions: %w", err)
		}

		for i, tx := range response.Data.SignedTxs {
			fmt.Printf("Signed transaction %d:\n", i)
			fmt.Println(tx)
		}
		return nil
	}

	// Confirm submission
	if !(c.Bool(flags.YesFlag) || utils.Confirm(confirmMessage)) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Submit them
	fmt.Println(submissionMessage)
	response, err := rp.Api.Tx.SubmitTxBatch(submissions, nonce, maxFee, maxPrioFee)
	if err != nil {
		return fmt.Errorf("error submitting transactions: %w", err)
	}

	// Wait for them
	utils.PrintTransactionBatchHashes(rp, response.Data.TxHashes)
	return waitForTransactions(rp, response.Data.TxHashes)
}

// Wait for a batch of transactions to get included in blocks
func waitForTransactions(rp *client.Client, hashes []common.Hash) error {
	var wg errgroup.Group
	var lock sync.Mutex
	total := len(hashes)
	successCount := 0

	// Create waiters for each TX
	for _, hash := range hashes {
		hash := hash
		wg.Go(func() error {
			if _, err := rp.Api.Tx.WaitForTransaction(hash); err != nil {
				return fmt.Errorf("error waiting for transaction %s: %w", hash.Hex(), err)
			}
			lock.Lock()
			successCount++
			fmt.Printf("Transaction %s complete (%d/%d)\n", hash.Hex(), successCount, total)
			lock.Unlock()
			return nil
		})
	}
	if err := wg.Wait(); err != nil {
		return fmt.Errorf("error waiting for transactions: %w", err)
	}
	return nil
}
