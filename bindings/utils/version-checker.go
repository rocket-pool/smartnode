package utils

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/hashicorp/go-version"
	"github.com/rocket-pool/smartnode/bindings/network"
	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
)

func GetCurrentVersion(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*version.Version, error) {

	// Check for v1.3.1 (Houston Hotfix)
	networkVotingVersion, err := network.GetRocketNetworkVotingVersion(rp, opts)
	if err != nil {
		return nil, fmt.Errorf("error checking network voting version: %w", err)
	}
	if networkVotingVersion > 1 {
		return version.NewSemver("1.3.1")
	}

	nodeMgrVersion, err := node.GetNodeManagerVersion(rp, opts)
	if err != nil {
		return nil, fmt.Errorf("error checking node manager version: %w", err)
	}

	// Check for v1.3 (Houston)
	if nodeMgrVersion > 3 {
		return version.NewSemver("1.3.0")
	}

	// Check for v1.2 (Atlas)
	nodeStakingVersion, err := node.GetNodeStakingVersion(rp, opts)
	if err != nil {
		return nil, fmt.Errorf("error checking node staking version: %w", err)
	}
	if nodeStakingVersion > 3 {
		return version.NewSemver("1.2.0")
	}

	// Check for v1.1 (Redstone)
	if nodeMgrVersion > 1 {
		return version.NewSemver("1.1.0")
	}

	// v1.0 (Classic)
	return version.NewSemver("1.0.0")

}
