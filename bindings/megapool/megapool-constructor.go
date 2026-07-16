package megapool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
)

// Create a megapool binding for the contract version deployed at the given address
func NewMegapool(rp *rocketpool.RocketPool, address common.Address, opts *bind.CallOpts) (Megapool, error) {

	// Get the contract version
	version, err := rocketpool.GetContractVersion(rp, address, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting megapool contract version: %w", err)
	}

	switch version {
	case 1:
		return NewMegaPoolV1(rp, address, opts)
	case 2:
		return NewMegaPoolV2(rp, address, opts)
	default:
		return nil, fmt.Errorf("unexpected megapool contract version [%d]", version)
	}
}
