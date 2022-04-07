package rp

import (
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

// Check if the contract upgrades for Merge support (Rocket Pool v1.1) have been deployed
func IsMergeUpdateDeployed(rp *rocketpool.RocketPool) (bool, error) {
	// Use rocketNodeManager's version as the reference
	version, err := node.GetNodeManagerVersion(rp, nil)
	if err != nil {
		return false, err
	}

	if version == 1 {
		return false, nil
	}

	return true, nil
}
