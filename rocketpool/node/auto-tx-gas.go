package node

import (
	"math/big"

	log "github.com/rocket-pool/smartnode/shared/logger"
	"github.com/rocket-pool/smartnode/shared/math"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rpgas "github.com/rocket-pool/smartnode/shared/services/gas"
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
		maxFee = math.GweiToWei(maxFeeGwei)
	}

	priorityFeeGwei := cfg.Smartnode.PriorityFee.Value.(float64)
	var maxPriorityFee *big.Int
	if priorityFeeGwei == 0 {
		logger.Printlnf("WARNING: priority fee was missing or 0, setting a default of %.2f.", rpgas.DefaultPriorityFeeGwei)
		maxPriorityFee = math.GweiToWei(rpgas.DefaultPriorityFeeGwei)
	} else {
		maxPriorityFee = math.GweiToWei(priorityFeeGwei)
	}

	maxFeeGweiLog := 0.0
	if maxFee != nil {
		maxFeeGweiLog = math.WeiToGwei(maxFee)
	}
	logger.Printlnf("Loaded auto-tx gas: threshold=%.4f gwei, maxFee=%.4f gwei (0=oracle), priorityFee=%.4f gwei",
		thresholdGwei, maxFeeGweiLog, math.WeiToGwei(maxPriorityFee))

	return autoTxGas{
		thresholdGwei:  thresholdGwei,
		maxFee:         maxFee,
		maxPriorityFee: maxPriorityFee,
	}
}
