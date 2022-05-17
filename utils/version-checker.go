package utils

import (
	"fmt"

	"github.com/hashicorp/go-version"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

func GetCurrentVersion(rp *rocketpool.RocketPool) (*version.Version, error) {

	// Base version
	rpVersion, _ := version.NewSemver("1.0.0")

	// Check for v1.1
	nodeMgrVersion, err := node.GetNodeManagerVersion(rp, nil)
	if err != nil {
		return nil, fmt.Errorf("Error checking node manager version: %w", err)
	}
	if nodeMgrVersion == 2 {
		rpVersion, _ = version.NewSemver("1.1.0")
	}

	// Return whatever version was found
	return rpVersion, nil

}
