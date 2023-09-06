package minipool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getMinipoolDelegateDetailsForNode(c *cli.Context) (*api.GetMinipoolDelegateDetailsForNodeResponse, error) {
	var delegate *core.Contract

	return runMinipoolQuery(c, MinipoolQuerier[api.GetMinipoolDelegateDetailsForNodeResponse]{
		CreateBindings: func(rp *rocketpool.RocketPool) error {
			var err error
			delegate, err = rp.GetContract(rocketpool.ContractName_RocketMinipoolDelegate)
			if err != nil {
				return fmt.Errorf("error getting minipool delegate binding: %w", err)
			}
			return nil
		},
		GetState:   nil,
		CheckState: nil,
		GetMinipoolDetails: func(mc *batch.MultiCaller, mp minipool.Minipool) {
			mpCommon := mp.GetMinipoolCommon()
			mpCommon.GetDelegate(mc)
			mpCommon.GetEffectiveDelegate(mc)
			mpCommon.GetPreviousDelegate(mc)
			mpCommon.GetUseLatestDelegate(mc)
		},
		PrepareResponse: func(rp *rocketpool.RocketPool, addresses []common.Address, mps []minipool.Minipool, response *api.GetMinipoolDelegateDetailsForNodeResponse) error {
			// Get all of the unique delegate addresses used by this node
			delegateAddresses := []common.Address{}
			delegateAddressMap := map[common.Address]bool{}
			for _, mp := range mps {
				mpCommonDetails := mp.GetMinipoolCommon().Details
				delegateAddressMap[mpCommonDetails.DelegateAddress] = true
				delegateAddressMap[mpCommonDetails.EffectiveDelegateAddress] = true
				delegateAddressMap[mpCommonDetails.PreviousDelegateAddress] = true
			}
			for delegateAddress := range delegateAddressMap {
				delegateAddresses = append(delegateAddresses, delegateAddress)
			}

			// Get the versions of each one
			versions := make([]uint8, len(delegateAddresses))
			delegateVersionMap := map[common.Address]uint8{}
			err := rp.Query(func(mc *batch.MultiCaller) error {
				for i, address := range delegateAddresses {
					err := rocketpool.GetContractVersion(mc, &versions[i], address)
					if err != nil {
						return fmt.Errorf("error getting version for delegate %s: %w", address.Hex(), err)
					}
				}
				return nil
			}, nil)
			if err != nil {
				return fmt.Errorf("error getting delegate versions: %w", err)
			}
			for i, address := range delegateAddresses {
				delegateVersionMap[address] = versions[i]
			}

			// Assign the details
			details := make([]api.MinipoolDelegateDetails, len(mps))
			for i, mp := range mps {
				mpCommonDetails := mp.GetMinipoolCommon().Details
				details[i] = api.MinipoolDelegateDetails{
					Address:           mpCommonDetails.Address,
					Delegate:          mpCommonDetails.DelegateAddress,
					EffectiveDelegate: mpCommonDetails.EffectiveDelegateAddress,
					PreviousDelegate:  mpCommonDetails.PreviousDelegateAddress,
					UseLatestDelegate: mpCommonDetails.IsUseLatestDelegateEnabled,
					RollbackVersionTooLow: (mpCommonDetails.DepositType.Formatted() == rptypes.Variable &&
						mpCommonDetails.Version >= 3 &&
						delegateVersionMap[mpCommonDetails.PreviousDelegateAddress] < 3),
					VersionTooLowToDisableUseLatest: (mpCommonDetails.DepositType.Formatted() == rptypes.Variable &&
						mpCommonDetails.Version >= 3 &&
						delegateVersionMap[mpCommonDetails.DelegateAddress] < 3),
				}
			}

			response.LatestDelegate = *delegate.Address
			response.Details = details
			return nil
		},
	})
}

func upgradeDelegates(c *cli.Context, minipoolAddresses []common.Address) (*api.BatchTxResponse, error) {
	return createBatchTxResponseForCommon(c, minipoolAddresses, func(mpCommon *minipool.MinipoolCommon, opts *bind.TransactOpts) (*core.TransactionInfo, error) {
		return mpCommon.DelegateUpgrade(opts)
	}, "upgrade-delegate")
}

func rollbackDelegates(c *cli.Context, minipoolAddresses []common.Address) (*api.BatchTxResponse, error) {
	return createBatchTxResponseForCommon(c, minipoolAddresses, func(mpCommon *minipool.MinipoolCommon, opts *bind.TransactOpts) (*core.TransactionInfo, error) {
		return mpCommon.DelegateRollback(opts)
	}, "rollback-delegate")
}

func setUseLatestDelegates(c *cli.Context, minipoolAddresses []common.Address, setting bool) (*api.BatchTxResponse, error) {
	return createBatchTxResponseForCommon(c, minipoolAddresses, func(mpCommon *minipool.MinipoolCommon, opts *bind.TransactOpts) (*core.TransactionInfo, error) {
		return mpCommon.SetUseLatestDelegate(setting, opts)
	}, "set-use-latest-delegate")
}
