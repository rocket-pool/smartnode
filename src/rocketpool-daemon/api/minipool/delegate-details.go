package minipool

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/core"
	"github.com/rocket-pool/rocketpool-go/v2/minipool"
	"github.com/rocket-pool/rocketpool-go/v2/node"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/v2/types"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
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

func (f *minipoolDelegateDetailsContextFactory) RegisterRoute(router *mux.Router) {
	RegisterMinipoolRoute[*minipoolDelegateDetailsContext, api.MinipoolDelegateDetailsData](
		router, "delegate/details", f, f.handler.ctx, f.handler.logger, f.handler.serviceProvider,
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

func (c *minipoolDelegateDetailsContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()

	// Bindings
	var err error
	c.delegate, err = c.rp.GetContract(rocketpool.ContractName_RocketMinipoolDelegate)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting minipool delegate binding: %w", err)
	}
	return types.ResponseStatus_Success, nil
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

func (c *minipoolDelegateDetailsContext) PrepareData(addresses []common.Address, mps []minipool.IMinipool, data *api.MinipoolDelegateDetailsData) (types.ResponseStatus, error) {
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
		return types.ResponseStatus_Error, fmt.Errorf("error getting delegate versions: %w", err)
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
	return types.ResponseStatus_Success, nil
}
