package rp

import (
	"github.com/hashicorp/go-version"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils"
)

// Check if the contract upgrades for Merge support (Rocket Pool v1.1) have been deployed
func IsMergeUpdateDeployed(rp *rocketpool.RocketPool) (bool, error) {
	currentVersion, err := utils.GetCurrentVersion(rp)
	if err != nil {
		return false, err
	}

	constraint, _ := version.NewConstraint(">= 1.1.0")
	return constraint.Check(currentVersion), nil
}
