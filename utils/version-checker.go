package utils

import (
	"fmt"

	"github.com/hashicorp/go-version"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

func GetCurrentVersion(rp *rocketpool.RocketPool) (*version.Version, error) {

	// Check for v1.2
	nodeStakingVersion, err := node.GetNodeStakingVersion(rp, nil)
	if err != nil {
		return nil, fmt.Errorf("error checking node staking version: %w", err)
	}
	if nodeStakingVersion == 3 {
		return version.NewSemver("1.2.0")
	}

	// Check for v1.1
	nodeMgrVersion, err := node.GetNodeManagerVersion(rp, nil)
	if err != nil {
		return nil, fmt.Errorf("error checking node manager version: %w", err)
	}
	if nodeMgrVersion == 2 {
		return version.NewSemver("1.1.0")
	}

	// v1.0
	return version.NewSemver("1.0.0")

}
