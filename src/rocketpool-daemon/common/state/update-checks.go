package state

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/hashicorp/go-version"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
)

// Check if Redstone has been deployed
func IsRedstoneDeployed(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	currentVersion, err := rp.GetProtocolVersion(opts)
	if err != nil {
		return false, err
	}

	constraint, _ := version.NewConstraint(">= 1.1.0")
	return constraint.Check(currentVersion), nil
}

// Check if Atlas has been deployed
func IsAtlasDeployed(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	currentVersion, err := rp.GetProtocolVersion(opts)
	if err != nil {
		return false, err
	}

	constraint, _ := version.NewConstraint(">= 1.2.0")
	return constraint.Check(currentVersion), nil
}

// Check if Houston has been deployed
func IsHoustonDeployed(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	currentVersion, err := rp.GetProtocolVersion(opts)
	if err != nil {
		return false, err
	}

	constraint, _ := version.NewConstraint(">= 1.3.0")
	return constraint.Check(currentVersion), nil
}
