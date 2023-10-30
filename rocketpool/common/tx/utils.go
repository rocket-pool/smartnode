package tx

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/config"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

// The fraction of the timeout period to trigger overdue transactions
const TimeoutSafetyFactor time.Duration = 2

// Prints a TX's details to the logger and waits for it to validated.
func PrintAndWaitForTransaction(cfg *config.RocketPoolConfig, rp *rocketpool.RocketPool, logger *log.ColorLogger, txInfo *core.TransactionInfo, opts *bind.TransactOpts) error {
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
	logger.Println("Waiting for the transaction to be validated...")

	// Wait for the TX to be included in a block
	err = rp.WaitForTransaction(tx)
	if err != nil {
		return fmt.Errorf("error waiting for transaction: %w", err)
	}

	return nil
}

// True if a transaction is due and needs to bypass the gas threshold
func IsTransactionDue(rp *rocketpool.RocketPool, startTime time.Time, minipoolLaunchTimeout time.Duration) (bool, time.Duration, error) {
	dueTime := minipoolLaunchTimeout / TimeoutSafetyFactor
	isDue := time.Since(startTime) > dueTime
	timeUntilDue := time.Until(startTime.Add(dueTime))
	return isDue, timeUntilDue, nil
}
