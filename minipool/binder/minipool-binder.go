package binder

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/minipool"
	minipool_v2 "github.com/rocket-pool/rocketpool-go/minipool/v2"
	minipool_v3 "github.com/rocket-pool/rocketpool-go/minipool/v3"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

func NewMinipool(rp *rocketpool.RocketPool, address common.Address, opts *bind.CallOpts) (minipool.Minipool, error) {

	// Get the contract version
	version, err := rocketpool.GetContractVersion(rp, address, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool contract version: %w", err)
	}

	switch version {
	case 1, 2:
		return minipool_v2.NewMinipool(rp, address, opts)
	case 3:
		return minipool_v3.NewMinipool(rp, address, opts)
	default:
		return nil, fmt.Errorf("unexpected minipool contract version [%d]", version)
	}
}
