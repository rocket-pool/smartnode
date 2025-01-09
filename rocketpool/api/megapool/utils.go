package megapool

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/megapool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"golang.org/x/sync/errgroup"
)

// Get all node minipool details
func GetNodeMegapoolDetails(rp *rocketpool.RocketPool, bc beacon.Client, nodeAddress common.Address) (api.MegapoolDetails, error) {

	// Get the megapool address
	var wg errgroup.Group

	details := api.MegapoolDetails{}

	wg.Go(func() error {
		var err error
		details.MegapoolAddress, err = megapool.GetMegapoolExpectedAddress(rp, nodeAddress, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		details.MegapoolDeployed, err = megapool.GetMegapoolDeployed(rp, nodeAddress, nil)
		return err
	})

	return details, nil
}
