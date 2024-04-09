package tx

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/gas"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/terminal"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

// Handle a transaction, either printing its details, signing it, or submitting it and waiting for it to be included
func HandleTx(c *cli.Context, rp *client.Client, txInfo *eth.TransactionInfo, confirmMessage string, identifier string, submissionMessage string) (bool, error) {
	// Print the TX data if requested
	if c.Bool(utils.PrintTxDataFlag.Name) {
		fmt.Printf("TX Data for %s:\n", identifier)
		fmt.Printf("\tTo:       %s\n", txInfo.To.Hex())
		fmt.Printf("\tData:     %s\n", hexutil.Encode(txInfo.Data))
		fmt.Printf("\tValue:    %s\n", txInfo.Value.String())
		fmt.Printf("\tEst. Gas: %d\n", txInfo.SimulationResult.EstimatedGasLimit)
		fmt.Printf("\tSafe Gas: %d\n", txInfo.SimulationResult.SafeGasLimit)

		// Warn if the TX failed simulation
		if txInfo.SimulationResult.SimulationError != "" {
			fmt.Printf("%sWARNING: '%s' failed simulation: %s\nThis transaction will likely revert if you submit it.%s\n", terminal.ColorYellow, identifier, txInfo.SimulationResult.SimulationError, terminal.ColorReset)
		}
		return false, nil
	}

	// Make sure the TX was successful
	if txInfo.SimulationResult.SimulationError != "" {
		return false, fmt.Errorf("simulating %s failed: %s", identifier, txInfo.SimulationResult.SimulationError)
	}

	// Assign max fees
	maxFee, maxPrioFee, err := gas.GetMaxFees(c, rp, txInfo.SimulationResult)
	if err != nil {
		return false, fmt.Errorf("error getting fee information: %w", err)
	}

	// Check the nonce flag
	var nonce *big.Int
	if rp.Context.Nonce.Cmp(common.Big0) > 0 {
		nonce = rp.Context.Nonce
	}

	// Create the submission from the TX info
	submission, _ := eth.CreateTxSubmissionFromInfo(txInfo, nil)

	// Sign only (no submission) if requested
	if c.Bool(utils.SignTxOnlyFlag.Name) {
		response, err := rp.Api.Tx.SignTx(submission, nonce, maxFee, maxPrioFee)
		if err != nil {
			return false, fmt.Errorf("error signing transaction: %w", err)
		}
		fmt.Printf("Signed transaction (%s):\n", identifier)
		fmt.Println(response.Data.SignedTx)
		fmt.Println()
		updateCustomNonce(rp)
		return false, nil
	}

	// Confirm submission
	if !(c.Bool(utils.YesFlag.Name) || utils.Confirm(confirmMessage)) {
		fmt.Println("Cancelled.")
		return false, nil
	}

	// Submit it
	fmt.Println(submissionMessage)
	response, err := rp.Api.Tx.SubmitTx(submission, nonce, maxFee, maxPrioFee)
	if err != nil {
		return false, fmt.Errorf("error submitting transaction: %w", err)
	}

	// Wait for it
	utils.PrintTransactionHash(rp, response.Data.TxHash)
	if _, err = rp.Api.Tx.WaitForTransaction(response.Data.TxHash); err != nil {
		return false, fmt.Errorf("error waiting for transaction: %w", err)
	}

	updateCustomNonce(rp)
	return true, nil
}

// Handle a batch of transactions, either printing their details, signing them, or submitting them and waiting for them to be included
func HandleTxBatch(c *cli.Context, rp *client.Client, txInfos []*eth.TransactionInfo, confirmMessage string, identifierFunc func(int) string, submissionMessage string) (bool, error) {
	// Print the TX data if requested
	if c.Bool(utils.PrintTxDataFlag.Name) {
		for i, info := range txInfos {
			id := identifierFunc(i)
			fmt.Printf("Data for TX %d (%s):\n", i, id)
			fmt.Printf("\tTo:       %s\n", info.To.Hex())
			fmt.Printf("\tData:     %s\n", hexutil.Encode(info.Data))
			fmt.Printf("\tValue:    %s\n", info.Value.String())
			fmt.Printf("\tEst. Gas: %d\n", info.SimulationResult.EstimatedGasLimit)
			fmt.Printf("\tSafe Gas: %d\n", info.SimulationResult.SafeGasLimit)
			fmt.Println()

			// Warn if the TX failed simulation
			if info.SimulationResult.SimulationError != "" {
				fmt.Printf("%sWARNING: '%s' failed simulation: %s\nThis transaction will likely revert if you submit it.%s\n", terminal.ColorYellow, id, info.SimulationResult.SimulationError, terminal.ColorReset)
				fmt.Println()
			}
		}
		return false, nil
	}

	// Make sure the TXs were successful
	for i, txInfo := range txInfos {
		if txInfo.SimulationResult.SimulationError != "" {
			return false, fmt.Errorf("simulating %s failed: %s", identifierFunc(i), txInfo.SimulationResult.SimulationError)
		}
	}

	// Assign max fees
	var gasInfo eth.SimulationResult
	for _, info := range txInfos {
		gasInfo.EstimatedGasLimit += info.SimulationResult.EstimatedGasLimit
		gasInfo.SafeGasLimit += info.SimulationResult.SafeGasLimit
	}
	maxFee, maxPrioFee, err := gas.GetMaxFees(c, rp, gasInfo)
	if err != nil {
		return false, fmt.Errorf("error getting fee information: %w", err)
	}

	// Check the nonce flag
	var nonce *big.Int
	if rp.Context.Nonce.Cmp(common.Big0) > 0 {
		nonce = rp.Context.Nonce
	}

	// Create the submissions from the TX infos
	submissions := make([]*eth.TransactionSubmission, len(txInfos))
	for i, info := range txInfos {
		submission, _ := eth.CreateTxSubmissionFromInfo(info, nil)
		submissions[i] = submission
	}

	// Sign only (no submission) if requested
	if c.Bool(utils.SignTxOnlyFlag.Name) {
		response, err := rp.Api.Tx.SignTxBatch(submissions, nonce, maxFee, maxPrioFee)
		if err != nil {
			return false, fmt.Errorf("error signing transactions: %w", err)
		}

		for i, tx := range response.Data.SignedTxs {
			fmt.Printf("Signed transaction (%s):\n", identifierFunc(i))
			fmt.Println(tx)
			fmt.Println()
		}
		return false, nil
	}

	// Confirm submission
	if !(c.Bool(utils.YesFlag.Name) || utils.Confirm(confirmMessage)) {
		fmt.Println("Cancelled.")
		return false, nil
	}

	// Submit them
	fmt.Println(submissionMessage)
	response, err := rp.Api.Tx.SubmitTxBatch(submissions, nonce, maxFee, maxPrioFee)
	if err != nil {
		return false, fmt.Errorf("error submitting transactions: %w", err)
	}

	// Wait for them
	utils.PrintTransactionBatchHashes(rp, response.Data.TxHashes)
	return true, waitForTransactions(rp, response.Data.TxHashes, identifierFunc)
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
