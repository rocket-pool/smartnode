package state

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/hashicorp/go-version"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils"
)

// Check if Redstone has been deployed
func IsRedstoneDeployed(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	currentVersion, err := utils.GetCurrentVersion(rp, opts)
	if err != nil {
		return false, err
	}

	constraint, _ := version.NewConstraint(">= 1.1.0")
	return constraint.Check(currentVersion), nil
}

// Check if Atlas has been deployed
func IsAtlasDeployed(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	currentVersion, err := utils.GetCurrentVersion(rp, opts)
	if err != nil {
		return false, err
	}

	constraint, _ := version.NewConstraint(">= 1.2.0")
	return constraint.Check(currentVersion), nil
}

// Check if Houston has been deployed
func IsHoustonDeployed(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	currentVersion, err := utils.GetCurrentVersion(rp, opts)
	if err != nil {
		return false, err
	}

	constraint, _ := version.NewConstraint(">= 1.3.0")
	return constraint.Check(currentVersion), nil
}

// Check if Houston Hotfix has been deployed
func IsHoustonHotfixDeployed(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	currentVersion, err := utils.GetCurrentVersion(rp, opts)
	if err != nil {
		return false, err
	}

	constraint, _ := version.NewConstraint(">= 1.3.1")
	return constraint.Check(currentVersion), nil
}
