package tx

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/rocket-pool/smartnode/v2/shared/keys"
	"golang.org/x/sync/errgroup"
)

// The fraction of the timeout period to trigger overdue transactions
const TimeoutSafetyFactor time.Duration = 2

// Prints a TX's details to the logger and waits for it to validated.
func PrintAndWaitForTransaction(cfg *config.SmartNodeConfig, rp *rocketpool.RocketPool, logger *slog.Logger, txInfo *eth.TransactionInfo, opts *bind.TransactOpts) error {
	tx, err := rp.ExecuteTransaction(txInfo, opts)
	if err != nil {
		return fmt.Errorf("error submitting transaction: %w", err)
	}

	resources := cfg.GetRocketPoolResources()
	txWatchUrl := resources.TxWatchUrl
	hashString := tx.Hash().String()
	logger.Info("Transaction has been submitted.", slog.String(keys.HashKey, hashString))
	if txWatchUrl != "" {
		logger.Info("You may follow its progress by visiting:")
		logger.Info(fmt.Sprintf("%s/%s\n", txWatchUrl, hashString))
	}

	// Wait for the TX to be included in a block
	logger.Info("Waiting for the transaction to be validated...")
	err = rp.WaitForTransaction(tx)
	if err != nil {
		return fmt.Errorf("error waiting for transaction: %w", err)
	}

	return nil
}

// Prints a TX's details to the logger and waits for it to validated.
func PrintAndWaitForTransactionBatch(cfg *config.SmartNodeConfig, rp *rocketpool.RocketPool, logger *slog.Logger, submissions []*eth.TransactionSubmission, callbacks []func(err error), opts *bind.TransactOpts) error {
	txs, err := rp.BatchExecuteTransactions(submissions, opts)
	if err != nil {
		return fmt.Errorf("error submitting transactions: %w", err)
	}

	// Make a map of hashes to callbacks
	callbackMap := map[common.Hash]func(err error){}
	if callbacks != nil {
		for i, tx := range txs {
			callbackMap[tx.Hash()] = callbacks[i]
		}
	}

	resources := cfg.GetRocketPoolResources()
	txWatchUrl := resources.TxWatchUrl
	if txWatchUrl != "" {
		logger.Info("Transactions have been submitted. You may follow them progress by visiting:")
		for _, tx := range txs {
			hashString := tx.Hash().String()
			logger.Info(fmt.Sprintf("%s/%s\n", txWatchUrl, hashString))
		}
	} else {
		logger.Info("Transactions have been submitted with the following hashes:")
		for _, tx := range txs {
			logger.Info(tx.Hash().String())
		}

	}

	// Wait for the TX to be included in a block
	logger.Info("Waiting for the transactions to be validated...")
	var wg errgroup.Group
	var waitLock sync.Mutex
	completeCount := 0

	for _, tx := range txs {
		tx := tx
		wg.Go(func() error {
			err := rp.WaitForTransaction(tx)
			if err != nil {
				err = fmt.Errorf("error waiting for transaction %s: %w", tx.Hash().String(), err)
			} else {
				waitLock.Lock()
				completeCount++
				logger.Info(fmt.Sprintf("TX %s complete (%d/%d)", tx.Hash().String(), completeCount, len(txs)))
				waitLock.Unlock()
			}

			// Run the callback with the error
			if callbacks != nil {
				callbackMap[tx.Hash()](err)
			}
			return err
		})
	}

	err = wg.Wait()
	if err != nil {
		return err
	}

	logger.Info("Transaction batch complete.")
	return nil
}

// True if a transaction is due and needs to bypass the gas threshold
func IsTransactionDue(startTime time.Time, minipoolLaunchTimeout time.Duration) (bool, time.Duration) {
	dueTime := minipoolLaunchTimeout / TimeoutSafetyFactor
	isDue := time.Since(startTime) > dueTime
	timeUntilDue := time.Until(startTime.Add(dueTime))
	return isDue, timeUntilDue
}
