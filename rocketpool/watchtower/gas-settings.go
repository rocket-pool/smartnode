package watchtower

import "github.com/rocket-pool/smartnode/shared/services/config"

const (
	minWatchtowerMaxFee      float64 = 200
	minWatchtowerPriorityFee float64 = 3
)

// Get the max fee for watchtower transactions
func getWatchtowerMaxFee(cfg *config.RocketPoolConfig) float64 {
	setting := cfg.Smartnode.WatchtowerMaxFeeOverride.Value.(float64)
	if setting < minWatchtowerMaxFee {
		return minWatchtowerMaxFee
	}
	return setting
}

// Get the priority fee for watchtower transactions
func getWatchtowerPrioFee(cfg *config.RocketPoolConfig) float64 {
	setting := cfg.Smartnode.WatchtowerPrioFeeOverride.Value.(float64)
	if setting < minWatchtowerPriorityFee {
		return minWatchtowerPriorityFee
	}
	return setting
}
