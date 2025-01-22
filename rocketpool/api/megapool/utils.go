package megapool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/megapool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"golang.org/x/sync/errgroup"
)

// Get all node megapool details
func GetNodeMegapoolDetails(rp *rocketpool.RocketPool, bc beacon.Client, nodeAccount common.Address) (api.MegapoolDetails, error) {

	megapoolAddress, err := megapool.GetMegapoolExpectedAddress(rp, nodeAccount, nil)
	if err != nil {
		return api.MegapoolDetails{}, err
	}

	// Load the megapool contract
	mega, err := megapool.NewMegaPoolV1(rp, megapoolAddress, nil)
	if err != nil {
		return api.MegapoolDetails{}, err
	}

	// Sync
	var wg errgroup.Group
	details := api.MegapoolDetails{Address: megapoolAddress}

	wg.Go(func() error {
		var err error
		details.NodeDebt, err = mega.GetDebt(nil)
		return err
	})
	wg.Go(func() error {
		var err error
		details.RefundValue, err = mega.GetRefundValue(nil)
		return err
	})
	wg.Go(func() error {
		var err error
		details.ValidatorCount, err = mega.GetValidatorCount(nil)
		return err
	})
	wg.Go(func() error {
		var err error
		details.PendingRewards, err = mega.GetPendingRewards(nil)
		return err
	})
	wg.Go(func() error {
		var err error
		details.UseLatestDelegate, err = mega.GetUseLatestDelegate(nil)
		return err
	})
	wg.Go(func() error {
		var err error
		details.DelegateAddress, err = mega.GetDelegate(nil)
		details.DelegateExpiry, err = megapool.GetMegapoolDelegateExpiry(rp, details.DelegateAddress, nil)
		return err
	})
	wg.Go(func() error {
		var err error
		details.EffectiveDelegateAddress, err = mega.GetEffectiveDelegate(nil)
		return err
	})
	wg.Go(func() error {
		var err error
		details.NodeExpressTicketCount, err = node.GetExpressTicketCount(rp, nodeAccount, nil)
		return err
	})
	wg.Go(func() error {
		var err error
		details.Deployed, err = megapool.GetMegapoolDeployed(rp, nodeAccount, nil)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return details, err
	}

	details.Validators, err = GetMegapoolValidatorDetails(rp, mega, nodeAccount, uint32(details.ValidatorCount))
	if err != nil {
		return details, err
	}

	return details, nil
}

func GetMegapoolValidatorDetails(rp *rocketpool.RocketPool, mp megapool.Megapool, nodeAccount common.Address, validatorCount uint32) ([]megapool.ValidatorInfo, error) {

	details := []megapool.ValidatorInfo{}

	for i := uint32(0); i < validatorCount; i++ {
		validatorDetails, err := mp.GetValidatorInfo(i, nil)
		if err != nil {
			fmt.Printf("Error retrieving validator %d details: %v\n", i, err)
		}
		details = append(details, validatorDetails)
	}

	return details, nil
}
