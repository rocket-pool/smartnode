package gas

import (
	"fmt"
	"math/big"

	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/log"
	"github.com/rocket-pool/smartnode/shared/gas/etherchain"
	"github.com/rocket-pool/smartnode/shared/gas/etherscan"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

// Print the gas price and cost of a TX
func PrintAndCheckGasInfo(gasInfo core.GasInfo, checkThreshold bool, gasThresholdGwei float64, logger *log.ColorLogger, maxFeeWei *big.Int, gasLimit uint64) bool {
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
	logger.Printlnf("This transaction will use a max fee of %.6f Gwei, for a total of up to %.6f - %.6f ETH.",
		eth.WeiToGwei(maxFeeWei),
		math.RoundDown(eth.WeiToEth(totalGasWei), 6),
		math.RoundDown(eth.WeiToEth(totalSafeGasWei), 6))

	return true
}

// Print the gas price and cost of a TX batch
func PrintAndCheckGasInfoForBatch(submissions []*eth.TransactionSubmission, checkThreshold bool, gasThresholdGwei float64, logger *log.ColorLogger, maxFeeWei *big.Int) bool {
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
	totalEstGasWei := big.NewInt(0)
	totalAssignedGasWei := big.NewInt(0)
	for _, submission := range submissions {
		lowGas := big.NewInt(0).SetUint64(submission.TxInfo.GasInfo.EstGasLimit)
		highGas := big.NewInt(0).SetUint64(submission.GasLimit)
		lowGas.Mul(lowGas, maxFeeWei)
		highGas.Mul(highGas, maxFeeWei)
		totalEstGasWei.Add(totalEstGasWei, lowGas)
		totalAssignedGasWei.Add(totalAssignedGasWei, highGas)
	}
	logger.Printlnf("These transactions combined will use a max fee of %.6f Gwei, for a total of up to %.6f - %.6f ETH.",
		eth.WeiToGwei(maxFeeWei),
		math.RoundDown(eth.WeiToEth(totalEstGasWei), 6),
		math.RoundDown(eth.WeiToEth(totalAssignedGasWei), 6))

	return true
}

// Get the suggested max fee for service operations
func GetMaxFeeWeiForDaemon(logger *log.ColorLogger) (*big.Int, error) {
	etherchainData, err := etherchain.GetGasPrices()
	if err == nil {
		return etherchainData.RapidWei, nil
	}

	logger.Println("WARNING: couldn't get gas estimates from Etherchain - %s\nFalling back to Etherscan\n", err.Error())
	etherscanData, err := etherscan.GetGasPrices()
	if err == nil {
		return eth.GweiToWei(etherscanData.FastGwei), nil
	}

	return nil, fmt.Errorf("error getting gas price suggestions: %w", err)
}
