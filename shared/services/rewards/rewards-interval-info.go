package rewards

import (
	"fmt"

	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

type rewardsIntervalInfo struct {
	rewardsRulesetVersion uint64
	mainnetStartInterval  uint64
	devnetStartInterval   uint64
	testnetStartInterval  uint64
	generator             treeGeneratorImpl
}

func (r *rewardsIntervalInfo) GetStartInterval(network cfgtypes.Network) (uint64, error) {
	switch network {
	case cfgtypes.Network_Mainnet:
		return r.mainnetStartInterval, nil
	case cfgtypes.Network_Devnet:
		return r.devnetStartInterval, nil
	case cfgtypes.Network_Testnet:
		return r.testnetStartInterval, nil
	default:
		return 0, fmt.Errorf("unknown network: %s", string(network))
	}
}
