package minipool

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type minipoolDelegateDetailsContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolDelegateDetailsContextFactory) Create(args url.Values) (*minipoolDelegateDetailsContext, error) {
	c := &minipoolDelegateDetailsContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *minipoolDelegateDetailsContextFactory) GetCancelContext() context.Context {
	return f.handler.context
}

func (f *minipoolDelegateDetailsContextFactory) RegisterRoute(router *mux.Router) {
	RegisterMinipoolRoute[*minipoolDelegateDetailsContext, api.MinipoolDelegateDetailsData](
		router, "delegate/details", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolDelegateDetailsContext struct {
	handler  *MinipoolHandler
	rp       *rocketpool.RocketPool
	delegate *core.Contract
}

func (c *minipoolDelegateDetailsContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()

	// Requirements
	err := errors.Join(
		sp.RequireNodeRegistered(c.handler.context),
	)
	if err != nil {
		return err
	}

	// Bindings
	c.delegate, err = c.rp.GetContract(rocketpool.ContractName_RocketMinipoolDelegate)
	if err != nil {
		return fmt.Errorf("error getting minipool delegate binding: %w", err)
	}
	return nil
}

func (c *minipoolDelegateDetailsContext) GetState(node *node.Node, mc *batch.MultiCaller) {
}

func (c *minipoolDelegateDetailsContext) CheckState(node *node.Node, response *api.MinipoolDelegateDetailsData) bool {
	return true
}

func (c *minipoolDelegateDetailsContext) GetMinipoolDetails(mc *batch.MultiCaller, mp minipool.IMinipool, index int) {
	mpCommon := mp.Common()
	eth.AddQueryablesToMulticall(mc,
		mpCommon.DelegateAddress,
		mpCommon.EffectiveDelegateAddress,
		mpCommon.PreviousDelegateAddress,
		mpCommon.IsUseLatestDelegateEnabled,
	)
}

func (c *minipoolDelegateDetailsContext) PrepareData(addresses []common.Address, mps []minipool.IMinipool, data *api.MinipoolDelegateDetailsData) error {
	// Get all of the unique delegate addresses used by this node
	delegateAddresses := []common.Address{}
	delegateAddressMap := map[common.Address]bool{}
	for _, mp := range mps {
		mpCommonDetails := mp.Common()
		delegateAddressMap[mpCommonDetails.DelegateAddress.Get()] = true
		delegateAddressMap[mpCommonDetails.EffectiveDelegateAddress.Get()] = true
		delegateAddressMap[mpCommonDetails.PreviousDelegateAddress.Get()] = true
	}
	for delegateAddress := range delegateAddressMap {
		delegateAddresses = append(delegateAddresses, delegateAddress)
	}

	// Get the versions of each one
	versions := make([]uint8, len(delegateAddresses))
	delegateVersionMap := map[common.Address]uint8{}
	err := c.rp.Query(func(mc *batch.MultiCaller) error {
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
		mpCommonDetails := mp.Common()
		details[i] = api.MinipoolDelegateDetails{
			Address:           mpCommonDetails.Address,
			Delegate:          mpCommonDetails.DelegateAddress.Get(),
			EffectiveDelegate: mpCommonDetails.EffectiveDelegateAddress.Get(),
			PreviousDelegate:  mpCommonDetails.PreviousDelegateAddress.Get(),
			UseLatestDelegate: mpCommonDetails.IsUseLatestDelegateEnabled.Get(),
			RollbackVersionTooLow: (mpCommonDetails.DepositType.Formatted() == rptypes.Variable &&
				mpCommonDetails.Version >= 3 &&
				delegateVersionMap[mpCommonDetails.PreviousDelegateAddress.Get()] < 3),
			VersionTooLowToDisableUseLatest: (mpCommonDetails.DepositType.Formatted() == rptypes.Variable &&
				mpCommonDetails.Version >= 3 &&
				delegateVersionMap[mpCommonDetails.DelegateAddress.Get()] < 3),
		}
	}

	data.LatestDelegate = c.delegate.Address
	data.Details = details
	return nil
}
