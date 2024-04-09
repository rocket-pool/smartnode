package rewards

import (
	"fmt"

	"github.com/rocket-pool/node-manager-core/config"
	snCfg "github.com/rocket-pool/smartnode/v2/shared/config"
)

type rewardsIntervalInfo struct {
	rewardsRulesetVersion uint64
	mainnetStartInterval  uint64
	devnetStartInterval   uint64
	holeskyStartInterval  uint64
	generator             treeGeneratorImpl
}

func (r *rewardsIntervalInfo) GetStartInterval(network config.Network) (uint64, error) {
	switch network {
	case config.Network_Mainnet:
		return r.mainnetStartInterval, nil
	case snCfg.Network_Devnet:
		return r.devnetStartInterval, nil
	case config.Network_Holesky:
		return r.holeskyStartInterval, nil
	default:
		return 0, fmt.Errorf("unknown network: %s", string(network))
	}
}
