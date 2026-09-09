package node

import (
	"math/big"

	log "github.com/rocket-pool/smartnode/shared/logger"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rpgas "github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/units"
)

// autoTxGas is the node-task view of the Smartnode auto-tx gas settings.
type autoTxGas struct {
	thresholdGwei  float64
	maxFee         *big.Int
	maxPriorityFee *big.Int
}

// loadAutoTxGas reads the auto-tx gas threshold, manual max fee, and priority
// fee from cfg. A missing or zero priority fee is replaced with the default
// and logged as a warning. A zero max fee is returned as a nil *big.Int so
// callers can fall back to oracle pricing.
func loadAutoTxGas(cfg *config.RocketPoolConfig, logger *log.ColorLogger) autoTxGas {
	thresholdGwei := cfg.Smartnode.AutoTxGasThreshold.Value.(float64)

	maxFeeGwei := cfg.Smartnode.ManualMaxFee.Value.(float64)
	var maxFee *big.Int
	if maxFeeGwei != 0 {
		maxFee = units.GweiFromFloat(maxFeeGwei).ToWei().BigInt()
	}

	priorityFeeGwei := cfg.Smartnode.PriorityFee.Value.(float64)
	var maxPriorityFee *big.Int
	if priorityFeeGwei == 0 {
		logger.Printlnf("WARNING: priority fee was missing or 0, setting a default of %.2f.", rpgas.DefaultPriorityFeeGwei)
		maxPriorityFee = units.GweiFromFloat(rpgas.DefaultPriorityFeeGwei).ToWei().BigInt()
	} else {
		maxPriorityFee = units.GweiFromFloat(priorityFeeGwei).ToWei().BigInt()
	}

	return autoTxGas{
		thresholdGwei:  thresholdGwei,
		maxFee:         maxFee,
		maxPriorityFee: maxPriorityFee,
	}
}
