package state

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/hashicorp/go-version"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/utils"
)

// Check if Saturn 2 has been deployed
func IsSaturn2Deployed(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	currentVersion, err := utils.GetCurrentVersion(rp, opts)
	if err != nil {
		return false, err
	}

	constraint, _ := version.NewConstraint(">= 1.5.0")
	return constraint.Check(currentVersion), nil
}
