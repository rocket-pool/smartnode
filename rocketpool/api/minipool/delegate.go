package minipool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getMinipoolDelegateDetailsForNode(c *cli.Context) (*api.GetMinipoolDelegateDetailsForNodeResponse, error) {
	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, fmt.Errorf("error getting node account: %w", err)
	}

	// Response
	response := api.GetMinipoolDelegateDetailsForNodeResponse{}

	// Create the bindings
	node, err := node.NewNode(rp, nodeAccount.Address)
	if err != nil {
		return nil, fmt.Errorf("error creating node %s binding: %w", nodeAccount.Address.Hex(), err)
	}
	delegate, err := rp.GetContract(rocketpool.ContractName_RocketMinipoolDelegate)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool delegate binding: %w", err)
	}
	response.LatestDelegate = *delegate.Address

	// Get contract state
	err = rp.Query(func(mc *batch.MultiCaller) error {
		node.GetMinipoolCount(mc)
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting contract state: %w", err)
	}

	// Get the minipool addresses for this node
	addresses, err := node.GetMinipoolAddresses(node.Details.MinipoolCount.Formatted(), nil)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool addresses: %w", err)
	}

	// Create each minipool binding
	mps, err := minipool.CreateMinipoolsFromAddresses(rp, addresses, true, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool details: %w", err)
	}

	// Get delegate details
	err = rp.BatchQuery(len(addresses), minipoolCompleteShareBatchSize, func(mc *batch.MultiCaller, i int) error {
		mpCommon := mps[i].GetMinipoolCommon()
		mpCommon.GetDelegate(mc)
		mpCommon.GetEffectiveDelegate(mc)
		mpCommon.GetPreviousDelegate(mc)
		mpCommon.GetUseLatestDelegate(mc)
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool delegate info: %w", err)
	}

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
	versions := make([]uint8, len(addresses))
	delegateVersionMap := map[common.Address]uint8{}
	err = rp.Query(func(mc *batch.MultiCaller) error {
		for i, address := range delegateAddresses {
			err := rocketpool.GetContractVersion(mc, &versions[i], address)
			if err != nil {
				return fmt.Errorf("error getting version for delegate %s: %w", address.Hex(), err)
			}
		}
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting delegate versions: %w", err)
	}
	for i, address := range delegateAddresses {
		delegateVersionMap[address] = versions[i]
	}

	// Assign the details
	details := make([]api.MinipoolDelegateDetails, len(addresses))
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

	response.Details = details
	return &response, nil
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
