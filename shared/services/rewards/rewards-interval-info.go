package rewards

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/services/config"
)

type rewardsIntervalInfo struct {
	rewardsRulesetVersion uint64
	generator             treeGeneratorImpl
}

func (r *rewardsIntervalInfo) GetStartInterval(cfg *config.RocketPoolConfig) (uint64, error) {
	info := cfg.GetNetworkInfo()
	if info == nil {
		return 0, fmt.Errorf("unknown network: %s", cfg.GetNetwork())
	}
	return info.Rewards.StartInterval(r.rewardsRulesetVersion), nil
}
