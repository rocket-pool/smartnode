package api

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	"github.com/rocket-pool/rocketpool-go/utils"
	"github.com/rocket-pool/rocketpool-go/utils/client"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

// The fraction of the timeout period to trigger overdue transactions
const TimeoutSafetyFactor int = 2

// Print the gas price and cost of a TX
func PrintAndCheckGasInfo(gasInfo rocketpool.GasInfo, checkThreshold bool, gasThresholdGwei float64, logger log.ColorLogger, maxFeeWei *big.Int, gasLimit uint64) bool {

	// Check the gas threshold if requested
	if checkThreshold {
		gasThresholdWei := math.RoundUp(gasThresholdGwei*eth.WeiPerGwei, 0)
		gasThreshold := new(big.Int).SetUint64(uint64(gasThresholdWei))
		if maxFeeWei.Cmp(gasThreshold) != -1 {
			logger.Printlnf("Current network gas price is %.2f Gwei, which is not lower than the set threshold of %.2f Gwei. "+
				"Aborting the transaction.", eth.WeiToGwei(maxFeeWei), gasThresholdGwei)
			return false
		}
	} else {
		logger.Println("This transaction does not check the gas threshold limit, continuing...")
	}

	// Print the total TX cost
	var gas *big.Int
	var safeGas *big.Int
	if gasLimit != 0 {
		gas = new(big.Int).SetUint64(gasLimit)
		safeGas = gas
	} else {
		gas = new(big.Int).SetUint64(gasInfo.EstGasLimit)
		safeGas = new(big.Int).SetUint64(gasInfo.SafeGasLimit)
	}
	totalGasWei := new(big.Int).Mul(maxFeeWei, gas)
	totalSafeGasWei := new(big.Int).Mul(maxFeeWei, safeGas)
	logger.Printlnf("This transaction will use a gas price of %.6f Gwei, for a total of %.6f to %.6f ETH.",
		eth.WeiToGwei(maxFeeWei),
		math.RoundDown(eth.WeiToEth(totalGasWei), 6),
		math.RoundDown(eth.WeiToEth(totalSafeGasWei), 6))

	return true
}

// Print a TX's details to the logger and waits for it to be mined.
func PrintAndWaitForTransaction(config config.RocketPoolConfig, hash common.Hash, ec *client.EthClientProxy, logger log.ColorLogger) error {

	txWatchUrl := config.Smartnode.TxWatchUrl
	hashString := hash.String()

	logger.Printlnf("Transaction has been submitted with hash %s.", hashString)
	if txWatchUrl != "" {
		logger.Printlnf("You may follow its progress by visiting:")
		logger.Printlnf("%s/%s\n", txWatchUrl, hashString)
	}
	logger.Println("Waiting for the transaction to be mined...")

	// Wait for the TX to be mined
	if _, err := utils.WaitForTransaction(ec, hash); err != nil {
		return fmt.Errorf("Error mining transaction: %w", err)
	}

	return nil

}

// Gets the event log interval supported by the selected eth1 client
func GetEventLogInterval(cfg config.RocketPoolConfig) (*big.Int, error) {

	// Get event log interval
	var eventLogInterval *big.Int = nil
	eth1Client := cfg.GetSelectedEth1Client()
	if eth1Client != nil && eth1Client.EventLogInterval != "" {
		var success bool
		eventLogInterval, success = big.NewInt(0).SetString(eth1Client.EventLogInterval, 0)
		if !success {
			return nil, fmt.Errorf("Cannot parse event log interval of [%s]", eth1Client.EventLogInterval)
		}
	}

	return eventLogInterval, nil

}

// True if a transaction is due and needs to bypass the gas threshold
func IsTransactionDue(rp *rocketpool.RocketPool, startTime time.Time) (bool, time.Duration, error) {

	// Get the dissolve timeout
	timeout, err := protocol.GetMinipoolLaunchTimeout(rp, nil)
	if err != nil {
		return false, 0, err
	}

	dueTime := timeout / time.Duration(TimeoutSafetyFactor)
	isDue := time.Since(startTime) > dueTime
	timeUntilDue := time.Until(startTime.Add(dueTime))
	return isDue, timeUntilDue, nil

}
