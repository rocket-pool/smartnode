package utils

import "github.com/rocket-pool/smartnode/shared/services/config"

const (
	MinWatchtowerMaxFee      float64 = 200
	MinWatchtowerPriorityFee float64 = 3
)

// Get the max fee for watchtower transactions
func GetWatchtowerMaxFee(cfg *config.RocketPoolConfig) float64 {
	setting := cfg.Smartnode.WatchtowerMaxFeeOverride.Value.(float64)
	if setting < MinWatchtowerMaxFee {
		return MinWatchtowerMaxFee
	}
	return setting
}

// Get the priority fee for watchtower transactions
func GetWatchtowerPrioFee(cfg *config.RocketPoolConfig) float64 {
	setting := cfg.Smartnode.WatchtowerPrioFeeOverride.Value.(float64)
	if setting < MinWatchtowerPriorityFee {
		return MinWatchtowerPriorityFee
	}
	return setting
}
