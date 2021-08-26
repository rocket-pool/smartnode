package api

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

// Print the gas price and cost of a TX
func PrintAndCheckGasInfo(gasInfo rocketpool.GasInfo, checkThreshold bool, gasThreshold float64, logger log.ColorLogger) (bool) {

    // Use the requested gas price if provided
    gasPrice := gasInfo.ReqGasPrice
    if gasPrice == nil {
        gasPrice = gasInfo.EstGasPrice
    }

    // Check the gas threshold if requested
    if checkThreshold {
        gasThresholdGwei := math.RoundUp(gasThreshold * eth.WeiPerGwei, 0)
        gasThreshold := new(big.Int).SetUint64(uint64(gasThresholdGwei))
        if gasPrice.Cmp(gasThreshold) != -1 {
            logger.Printlnf("Current network gas price is %.6f Gwei, which is higher than the set threshold of %.6f Gwei. " + 
                "Aborting the transaction.", eth.WeiToGwei(gasInfo.EstGasPrice), eth.WeiToGwei(gasThreshold))
            return false
        } 
    } else {
        logger.Println("This transaction does not check the gas threshold limit, continuing...")
    }
    
    // Print the total TX cost
    var gas *big.Int 
    var safeGas *big.Int 
    if gasInfo.ReqGasLimit != 0 {
        gas = new(big.Int).SetUint64(gasInfo.ReqGasLimit)
        safeGas = gas
    } else {
        gas = new(big.Int).SetUint64(gasInfo.EstGasLimit)
        safeGas = new(big.Int).SetUint64(gasInfo.SafeGasLimit)
    }
    totalGasWei := new(big.Int).Mul(gasPrice, gas)
    totalSafeGasWei := new(big.Int).Mul(gasPrice, safeGas)
    logger.Printlnf("This transaction will use a gas price of %.6f Gwei, for a total of %.6f to %.6f ETH.",
        eth.WeiToGwei(gasPrice),
        math.RoundDown(eth.WeiToEth(totalGasWei), 6),
        math.RoundDown(eth.WeiToEth(totalSafeGasWei), 6))
        
    return true
}


// Print a TX's details to the logger and waits for it to be mined.
func PrintAndWaitForTransaction(config config.RocketPoolConfig, hash common.Hash, ec *ethclient.Client, logger log.ColorLogger) (error) {

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

