package tx

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/client"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/gas"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

// Handle a transaction, either printing its details, signing it, or submitting it and waiting for it to be included
func HandleTx(c *cli.Context, rp *client.Client, txInfo *core.TransactionInfo, confirmMessage string, identifier string, submissionMessage string) error {
	// Make sure the TX was successful
	if txInfo.SimError != "" {
		return fmt.Errorf("simulating %s failed: %s", identifier, txInfo.SimError)
	}

	// Print the TX data if requested
	if c.Bool(utils.PrintTxDataFlag) {
		fmt.Printf("TX Data for %s:\n", identifier)
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
	if rp.Context.Nonce.Cmp(common.Big0) > 0 {
		nonce = rp.Context.Nonce
	}

	// Create the submission from the TX info
	submission, _ := core.CreateTxSubmissionFromInfo(txInfo, nil)

	// Sign only (no submission) if requested
	if c.Bool(utils.SignTxOnlyFlag) {
		response, err := rp.Api.Tx.SignTx(submission, nonce, maxFee, maxPrioFee)
		if err != nil {
			return fmt.Errorf("error signing transaction: %w", err)
		}
		fmt.Printf("Signed transaction (%s):\n", identifier)
		fmt.Println(response.Data.SignedTx)
		fmt.Println()
		updateCustomNonce(rp)
		return nil
	}

	// Confirm submission
	if !(c.Bool(utils.YesFlag.Name) || utils.Confirm(confirmMessage)) {
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

	updateCustomNonce(rp)
	return nil
}

// Handle a batch of transactions, either printing their details, signing them, or submitting them and waiting for them to be included
func HandleTxBatch(c *cli.Context, rp *client.Client, txInfos []*core.TransactionInfo, confirmMessage string, identifierFunc func(int) string, submissionMessage string) error {
	// Make sure the TXs were successful
	for i, txInfo := range txInfos {
		if txInfo.SimError != "" {
			return fmt.Errorf("simulating %s failed: %s", identifierFunc(i), txInfo.SimError)
		}
	}

	// Print the TX data if requested
	if c.Bool(utils.PrintTxDataFlag) {
		for i, info := range txInfos {
			fmt.Printf("Data for TX %d (%s):\n", i, identifierFunc(i))
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
	if rp.Context.Nonce.Cmp(common.Big0) > 0 {
		nonce = rp.Context.Nonce
	}

	// Create the submissions from the TX infos
	submissions := make([]*core.TransactionSubmission, len(txInfos))
	for i, info := range txInfos {
		submission, _ := core.CreateTxSubmissionFromInfo(info, nil)
		submissions[i] = submission
	}

	// Sign only (no submission) if requested
	if c.Bool(utils.SignTxOnlyFlag) {
		response, err := rp.Api.Tx.SignTxBatch(submissions, nonce, maxFee, maxPrioFee)
		if err != nil {
			return fmt.Errorf("error signing transactions: %w", err)
		}

		for i, tx := range response.Data.SignedTxs {
			fmt.Printf("Signed transaction (%s):\n", identifierFunc(i))
			fmt.Println(tx)
			fmt.Println()
		}
		return nil
	}

	// Confirm submission
	if !(c.Bool(utils.YesFlag.Name) || utils.Confirm(confirmMessage)) {
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
	return waitForTransactions(rp, response.Data.TxHashes, identifierFunc)
}

// Wait for a batch of transactions to get included in blocks
func waitForTransactions(rp *client.Client, hashes []common.Hash, identifierFunc func(int) string) error {
	var wg errgroup.Group
	var lock sync.Mutex
	total := len(hashes)
	successCount := 0

	// Create waiters for each TX
	for i, hash := range hashes {
		i := i
		hash := hash
		wg.Go(func() error {
			if _, err := rp.Api.Tx.WaitForTransaction(hash); err != nil {
				return fmt.Errorf("error waiting for transaction %s: %w", hash.Hex(), err)
			}
			lock.Lock()
			successCount++
			fmt.Printf("TX %s (%s) complete (%d/%d)\n", hash.Hex(), identifierFunc(i), successCount, total)
			lock.Unlock()
			return nil
		})
	}
	if err := wg.Wait(); err != nil {
		return fmt.Errorf("error waiting for transactions: %w", err)
	}
	return nil
}

// If a custom nonce is set, increment it for the next transaction
func updateCustomNonce(rp *client.Client) {
	if rp.Context.Nonce.Cmp(common.Big0) > 0 {
		rp.Context.Nonce.Add(rp.Context.Nonce, common.Big1)
	}
}
