package gaslimit

import (
	"math/big"

	log "github.com/rocket-pool/smartnode/shared/logger"
	"github.com/rocket-pool/smartnode/shared/units"
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

func (l Limits) Check(checkThreshold bool, gasThreshold units.Gwei, logger *log.ColorLogger, maxFee units.Wei, gasLimit uint64) bool {
	if !checkThreshold {
		logger.Println("This transaction does not check the gas threshold limit, continuing...")
		return true
	}

	if maxFee.Cmp(gasThreshold.ToWei()) != -1 {
		logger.Printlnf("Current network gas price is %.2f Gwei, which is not lower than the set threshold of %.2f Gwei. "+
			"Aborting the transaction.", maxFee.ToGwei(), gasThreshold)
		return false
	}

	return true
}

func (l Limits) Print(logger *log.ColorLogger, maxFee units.Wei, gasLimit uint64) {

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
	totalGasWei := new(big.Int).Mul(maxFee.BigInt(), gas)
	totalSafeGasWei := new(big.Int).Mul(maxFee.BigInt(), safeGas)
	logger.Printlnf("This transaction will use a max fee of %.6f Gwei, for a total of up to %.6f - %.6f ETH.",
		maxFee.ToGwei(),
		units.NewWei(totalGasWei).ToEth(),
		units.NewWei(totalSafeGasWei).ToEth(),
	)
}

func (l Limits) PrintAndCheck(checkThreshold bool, gasThreshold units.Gwei, logger *log.ColorLogger, maxFee units.Wei, gasLimit uint64) bool {
	if !l.Check(checkThreshold, gasThreshold, logger, maxFee, gasLimit) {
		return false
	}

	l.Print(logger, maxFee, gasLimit)
	return true
}
