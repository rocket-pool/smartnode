package tx

import (
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/log"
	"github.com/rocket-pool/smartnode/shared/config"
	"golang.org/x/sync/errgroup"
)

// The fraction of the timeout period to trigger overdue transactions
const TimeoutSafetyFactor time.Duration = 2

// Prints a TX's details to the logger and waits for it to validated.
func PrintAndWaitForTransaction(cfg *config.RocketPoolConfig, rp *rocketpool.RocketPool, logger *log.ColorLogger, txInfo *eth.TransactionInfo, opts *bind.TransactOpts) error {
	tx, err := rp.ExecuteTransaction(txInfo, opts)
	if err != nil {
		return fmt.Errorf("error submitting transaction: %w", err)
	}

	txWatchUrl := cfg.Smartnode.GetTxWatchUrl()
	hashString := tx.Hash().String()
	logger.Printlnf("Transaction has been submitted with hash %s.", hashString)
	if txWatchUrl != "" {
		logger.Printlnf("You may follow its progress by visiting:")
		logger.Printlnf("%s/%s\n", txWatchUrl, hashString)
	}

	// Wait for the TX to be included in a block
	logger.Println("Waiting for the transaction to be validated...")
	err = rp.WaitForTransaction(tx)
	if err != nil {
		return fmt.Errorf("error waiting for transaction: %w", err)
	}

	return nil
}

// Prints a TX's details to the logger and waits for it to validated.
func PrintAndWaitForTransactionBatch(cfg *config.RocketPoolConfig, rp *rocketpool.RocketPool, logger *log.ColorLogger, submissions []*eth.TransactionSubmission, opts *bind.TransactOpts) error {
	txs, err := rp.BatchExecuteTransactions(submissions, opts)
	if err != nil {
		return fmt.Errorf("error submitting transactions: %w", err)
	}

	txWatchUrl := cfg.Smartnode.GetTxWatchUrl()
	if txWatchUrl != "" {
		logger.Println("Transactions have been submitted. You may follow them progress by visiting:")
		for _, tx := range txs {
			hashString := tx.Hash().String()
			logger.Printlnf("%s/%s\n", txWatchUrl, hashString)
		}
	} else {
		logger.Println("Transactions have been submitted with the following hashes:")
		for _, tx := range txs {
			logger.Println(tx.Hash().String())
		}

	}

	// Wait for the TX to be included in a block
	logger.Println("Waiting for the transactions to be validated...")
	var wg errgroup.Group
	var waitLock sync.Mutex
	completeCount := 0

	for _, tx := range txs {
		tx := tx
		wg.Go(func() error {
			err := rp.WaitForTransaction(tx)
			if err != nil {
				return fmt.Errorf("error waiting for transaction %s: %w", tx.Hash().String(), err)
			}
			waitLock.Lock()
			completeCount++
			logger.Println("TX %s complete (%d/%d)", tx.Hash().String(), completeCount, len(txs))
			waitLock.Unlock()
			return nil
		})
	}

	err = wg.Wait()
	if err != nil {
		return err
	}

	logger.Println("Transaction batch complete.")
	return nil
}

// True if a transaction is due and needs to bypass the gas threshold
func IsTransactionDue(startTime time.Time, minipoolLaunchTimeout time.Duration) (bool, time.Duration) {
	dueTime := minipoolLaunchTimeout / TimeoutSafetyFactor
	isDue := time.Since(startTime) > dueTime
	timeUntilDue := time.Until(startTime.Add(dueTime))
	return isDue, timeUntilDue
}
