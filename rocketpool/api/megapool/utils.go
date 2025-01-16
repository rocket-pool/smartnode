package megapool

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/megapool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"golang.org/x/sync/errgroup"
)

// Get all node minipool details
func GetNodeMegapoolDetails(rp *rocketpool.RocketPool, bc beacon.Client, nodeAccount common.Address) (api.MegapoolDetails, error) {

	details := api.MegapoolDetails{}

	// Sync
	var wg errgroup.Group

	wg.Go(func() error {
		megapoolAddress, err := megapool.GetMegapoolExpectedAddress(rp, nodeAccount, nil)
		if err == nil {
			details.Address = megapoolAddress
		}

		// Load the megapool contract
		mp, err := megapool.NewMegaPoolV1(rp, megapoolAddress, nil)
		if err == nil {
			debt, err := mp.GetDebt(nil)
			if err == nil {
				details.NodeDebt = debt
			}
			refund, err := mp.GetRefundValue(nil)
			if err == nil {
				details.RefundValue = refund
			}
			validatorCount, err := mp.GetValidatorCount(nil)
			if err == nil {
				details.ValidatorCount = uint16(validatorCount)
			}
			pendingRewards, err := mp.GetPendingRewards(nil)
			if err == nil {
				details.PendingRewards = pendingRewards
			}
			useLatest, err := mp.GetUseLatestDelegate(nil)
			if err == nil {
				details.UseLatestDelegate = useLatest
			}
			delegateAddress, err := mp.GetDelegate(nil)
			if err == nil {
				details.DelegateAddress = delegateAddress
			}
			effectiveDelegateAddress, err := mp.GetEffectiveDelegate(nil)
			if err == nil {
				details.EffectiveDelegateAddress = effectiveDelegateAddress
			}
		}
		return err
	})

	wg.Go(func() error {
		expressTicketCount, err := node.GetExpressTicketCount(rp, nodeAccount, nil)
		if err == nil {
			details.NodeExpressTicketCount = expressTicketCount
		}
		return err
	})

	wg.Go(func() error {
		megapoolDepoyed, err := megapool.GetMegapoolDeployed(rp, nodeAccount, nil)
		if err == nil {
			details.Deployed = megapoolDepoyed
		}
		return err
	})

	// wg.Go(func() error {
	// 	megapoolDelegate, err := rp.GetContract("rocketMegapoolDelegate", nil)
	// 	if err == nil {
	// 		details.DelegateAddress = *megapoolDelegate.Address
	// 	}

	// 	megapoolDelegateExpiry, err := megapool.GetMegapoolDelegateExpiry(rp, details.DelegateAddress, nil)
	// 	if err == nil {
	// 		details.DelegateExpiry = megapoolDelegateExpiry
	// 	}
	// 	return err
	// })

	// Wait for data
	if err := wg.Wait(); err != nil {
		return details, err
	}

	return details, nil
}
