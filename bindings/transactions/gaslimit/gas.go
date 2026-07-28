package gaslimit

import (
	"math/big"

	log "github.com/rocket-pool/smartnode/shared/logger"
	"github.com/rocket-pool/smartnode/shared/math"
)

// Response for gas limits from network and from user request
type Limits struct {
	Estimated uint64 `json:"estimated"`
	Safe      uint64 `json:"safe"`
}

func (l Limits) IsBlank() bool {
	return l == Limits{}
}

func (l Limits) Add(other Limits) Limits {
	return Limits{
		Estimated: l.Estimated + other.Estimated,
		Safe:      l.Safe + other.Safe,
	}
}

func (l Limits) Check(checkThreshold bool, gasThresholdGwei float64, logger *log.ColorLogger, maxFeeWei *big.Int, gasLimit uint64) bool {
	if !checkThreshold {
		logger.Println("This transaction does not check the gas threshold limit, continuing...")
		return true
	}

	gasThresholdWei := math.RoundUp(gasThresholdGwei*math.WeiPerGwei, 0)
	gasThreshold := new(big.Int).SetUint64(uint64(gasThresholdWei))
	if maxFeeWei.Cmp(gasThreshold) != -1 {
		logger.Printlnf("Current network gas price is %.2f Gwei, which is not lower than the set threshold of %.2f Gwei. "+
			"Aborting the transaction.", math.WeiToGwei(maxFeeWei), gasThresholdGwei)
		return false
	}

	return true
}

func (l Limits) Print(logger *log.ColorLogger, maxFeeWei *big.Int, gasLimit uint64) {

	// Print the total TX cost
	var gas *big.Int
	var safeGas *big.Int
	if gasLimit != 0 {
		gas = new(big.Int).SetUint64(gasLimit)
		safeGas = gas
	} else {
		gas = new(big.Int).SetUint64(l.Estimated)
		safeGas = new(big.Int).SetUint64(l.Safe)
	}
	totalGasWei := new(big.Int).Mul(maxFeeWei, gas)
	totalSafeGasWei := new(big.Int).Mul(maxFeeWei, safeGas)
	logger.Printlnf("This transaction will use a max fee of %.6f Gwei, for a total of up to %.6f - %.6f ETH.",
		math.WeiToGwei(maxFeeWei),
		math.RoundDown(math.WeiToEth(totalGasWei), 6),
		math.RoundDown(math.WeiToEth(totalSafeGasWei), 6))
}

func (l Limits) PrintAndCheck(checkThreshold bool, gasThresholdGwei float64, logger *log.ColorLogger, maxFeeWei *big.Int, gasLimit uint64) bool {
	if !l.Check(checkThreshold, gasThresholdGwei, logger, maxFeeWei, gasLimit) {
		return false
	}

	l.Print(logger, maxFeeWei, gasLimit)
	return true
}
