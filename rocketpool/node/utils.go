package node

import (
	"math/big"

	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/config"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

func getAutoTxInfo(cfg *config.RocketPoolConfig, logger *log.ColorLogger) (bool, *big.Int, *big.Int, float64) {
	// Check if auto-distributing is disabled
	gasThreshold := cfg.Smartnode.AutoTxGasThreshold.Value.(float64)
	disabled := false
	if gasThreshold == 0 {
		logger.Println("Automatic tx gas threshold is 0, disabling auto-distribute.")
		disabled = true
	}

	// Get the user-requested max fee
	maxFeeGwei := cfg.Smartnode.ManualMaxFee.Value.(float64)
	var maxFee *big.Int
	if maxFeeGwei == 0 {
		maxFee = nil
	} else {
		maxFee = eth.GweiToWei(maxFeeGwei)
	}

	// Get the user-requested max fee
	priorityFeeGwei := cfg.Smartnode.PriorityFee.Value.(float64)
	var priorityFee *big.Int
	if priorityFeeGwei == 0 {
		logger.Println("WARNING: priority fee was missing or 0, setting a default of 2.")
		priorityFee = eth.GweiToWei(2)
	} else {
		priorityFee = eth.GweiToWei(priorityFeeGwei)
	}

	return disabled, maxFee, priorityFee, gasThreshold
}
