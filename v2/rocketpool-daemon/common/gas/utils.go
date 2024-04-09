package gas

import (
	"fmt"
	"log/slog"
	"math/big"

	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/gas"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/node-manager-core/utils/math"
)

const (
	CurrentMaxFeeKey   string = "currentMaxFee"
	ThresholdMaxFeeKey string = "thresholdMaxFee"
	MinCostKey         string = "minCost"
	MaxCostKey         string = "maxCost"
	TotalMinCostKey    string = "totalMinCost"
	TotalMaxCostKey    string = "totalMaxCost"
)

// Print the gas price and cost of a TX
func PrintAndCheckGasInfo(simResult eth.SimulationResult, checkThreshold bool, gasThresholdGwei float64, logger *slog.Logger, maxFeeWei *big.Int, gasLimit uint64) bool {
	// Check the gas threshold if requested
	if checkThreshold {
		gasThresholdWei := math.RoundUp(gasThresholdGwei*eth.WeiPerGwei, 0)
		gasThreshold := new(big.Int).SetUint64(uint64(gasThresholdWei))
		if maxFeeWei.Cmp(gasThreshold) != -1 {
			logger.Warn("Current network gas price is too high, aborting the transaction.", slog.Float64(CurrentMaxFeeKey, eth.WeiToGwei(maxFeeWei)), slog.Float64(ThresholdMaxFeeKey, gasThresholdGwei))
			return false
		}
	} else {
		logger.Info("This transaction does not check the gas threshold limit, continuing...")
	}

	// Print the total TX cost
	var gas *big.Int
	var safeGas *big.Int
	if gasLimit != 0 {
		gas = new(big.Int).SetUint64(gasLimit)
		safeGas = gas
	} else {
		gas = new(big.Int).SetUint64(simResult.EstimatedGasLimit)
		safeGas = new(big.Int).SetUint64(simResult.SafeGasLimit)
	}
	totalGasWei := new(big.Int).Mul(maxFeeWei, gas)
	totalSafeGasWei := new(big.Int).Mul(maxFeeWei, safeGas)
	logger.Info("Current network gas is low enough to proceed", slog.Float64(CurrentMaxFeeKey, eth.WeiToGwei(maxFeeWei)), slog.Float64(MinCostKey, eth.WeiToEth(totalGasWei)), slog.Float64(MaxCostKey, eth.WeiToEth(totalSafeGasWei)))
	return true
}

// Print the gas price and cost of a TX batch
func PrintAndCheckGasInfoForBatch(submissions []*eth.TransactionSubmission, checkThreshold bool, gasThresholdGwei float64, logger *slog.Logger, maxFeeWei *big.Int) bool {
	// Check the gas threshold if requested
	if checkThreshold {
		gasThresholdWei := math.RoundUp(gasThresholdGwei*eth.WeiPerGwei, 0)
		gasThreshold := new(big.Int).SetUint64(uint64(gasThresholdWei))
		if maxFeeWei.Cmp(gasThreshold) != -1 {
			logger.Warn("Current network gas price is too high, aborting the transaction.", slog.Float64(CurrentMaxFeeKey, eth.WeiToGwei(maxFeeWei)), slog.Float64(ThresholdMaxFeeKey, gasThresholdGwei))
			return false
		}
	} else {
		logger.Info("This transaction does not check the gas threshold limit, continuing...")
	}

	// Print the total TX cost
	totalEstGasWei := big.NewInt(0)
	totalAssignedGasWei := big.NewInt(0)
	for _, submission := range submissions {
		lowGas := big.NewInt(0).SetUint64(submission.TxInfo.SimulationResult.EstimatedGasLimit)
		highGas := big.NewInt(0).SetUint64(submission.GasLimit)
		lowGas.Mul(lowGas, maxFeeWei)
		highGas.Mul(highGas, maxFeeWei)
		totalEstGasWei.Add(totalEstGasWei, lowGas)
		totalAssignedGasWei.Add(totalAssignedGasWei, highGas)
	}

	logger.Info("Current network gas is low enough to proceed", slog.Float64(CurrentMaxFeeKey, eth.WeiToGwei(maxFeeWei)), slog.Float64(TotalMinCostKey, eth.WeiToEth(totalEstGasWei)), slog.Float64(TotalMaxCostKey, eth.WeiToEth(totalAssignedGasWei)))
	return true
}

// Get the suggested max fee for service operations
func GetMaxFeeWeiForDaemon(logger *slog.Logger) (*big.Int, error) {
	etherchainData, err := gas.GetEtherchainGasPrices()
	if err == nil {
		return etherchainData.RapidWei, nil
	}

	logger.Warn("Couldn't get gas estimates from Etherchain, falling back to Etherscan\n", log.Err(err))
	etherscanData, err := gas.GetEtherscanGasPrices()
	if err == nil {
		return eth.GweiToWei(etherscanData.FastGwei), nil
	}

	return nil, fmt.Errorf("error getting gas price suggestions: %w", err)
}
