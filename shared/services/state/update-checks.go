package state

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/hashicorp/go-version"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils"
)

// Check if Saturn has been deployed
func IsSaturnDeployed(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	currentVersion, err := utils.GetCurrentVersion(rp, opts)
	if err != nil {
		return false, err
	}

	constraint, _ := version.NewConstraint(">= 1.4.0")
	return constraint.Check(currentVersion), nil
}
