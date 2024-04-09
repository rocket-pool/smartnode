package minipool

import (
	"fmt"
	"math/big"
	"net/url"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/minipool"
	"github.com/rocket-pool/rocketpool-go/v2/node"
	rptypes "github.com/rocket-pool/rocketpool-go/v2/types"

	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type minipoolRescueDissolvedDetailsContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolRescueDissolvedDetailsContextFactory) Create(args url.Values) (*minipoolRescueDissolvedDetailsContext, error) {
	c := &minipoolRescueDissolvedDetailsContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *minipoolRescueDissolvedDetailsContextFactory) RegisterRoute(router *mux.Router) {
	RegisterMinipoolRoute[*minipoolRescueDissolvedDetailsContext, api.MinipoolRescueDissolvedDetailsData](
		router, "rescue-dissolved/details", f, f.handler.ctx, f.handler.logger, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolRescueDissolvedDetailsContext struct {
	handler *MinipoolHandler
	bc      beacon.IBeaconClient
}

func (c *minipoolRescueDissolvedDetailsContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.bc = sp.GetBeaconClient()

	// Requirements
	err := sp.RequireBeaconClientSynced(c.handler.ctx)
	if err != nil {
		return types.ResponseStatus_ClientsNotSynced, err
	}
	return types.ResponseStatus_Success, nil
}

func (c *minipoolRescueDissolvedDetailsContext) GetState(node *node.Node, mc *batch.MultiCaller) {
}

func (c *minipoolRescueDissolvedDetailsContext) CheckState(node *node.Node, response *api.MinipoolRescueDissolvedDetailsData) bool {
	return true
}

func (c *minipoolRescueDissolvedDetailsContext) GetMinipoolDetails(mc *batch.MultiCaller, mp minipool.IMinipool, index int) {
	mpCommon := mp.Common()
	eth.AddQueryablesToMulticall(mc,
		mpCommon.IsFinalised,
		mpCommon.Status,
		mpCommon.Pubkey,
	)
}

func (c *minipoolRescueDissolvedDetailsContext) PrepareData(addresses []common.Address, mps []minipool.IMinipool, data *api.MinipoolRescueDissolvedDetailsData) (types.ResponseStatus, error) {
	ctx := c.handler.ctx
	// Get the rescue details
	pubkeys := []beacon.ValidatorPubkey{}
	detailsMap := map[beacon.ValidatorPubkey]int{}
	details := make([]api.MinipoolRescueDissolvedDetails, len(addresses))
	for i, mp := range mps {
		mpCommon := mp.Common()
		mpDetails := api.MinipoolRescueDissolvedDetails{
			Address:       mpCommon.Address,
			MinipoolState: mpCommon.Status.Formatted(),
			IsFinalized:   mpCommon.IsFinalised.Get(),
		}

		if mpDetails.MinipoolState != rptypes.MinipoolStatus_Dissolved || mpDetails.IsFinalized {
			mpDetails.InvalidElState = true
		} else {
			pubkey := mpCommon.Pubkey.Get()
			pubkeys = append(pubkeys, pubkey)
			detailsMap[pubkey] = i
		}

		details[i] = mpDetails
	}

	// Get the statuses on Beacon
	beaconStatuses, err := c.bc.GetValidatorStatuses(ctx, pubkeys, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting validator statuses on Beacon: %w", err)
	}

	// Do a complete viability check
	for pubkey, beaconStatus := range beaconStatuses {
		i := detailsMap[pubkey]
		mpDetails := &details[i]
		mpDetails.BeaconState = beaconStatus.Status
		mpDetails.InvalidBeaconState = beaconStatus.Status != beacon.ValidatorState_PendingInitialized

		if !mpDetails.InvalidBeaconState {
			beaconBalanceGwei := big.NewInt(0).SetUint64(beaconStatus.Balance)
			mpDetails.BeaconBalance = big.NewInt(0).Mul(beaconBalanceGwei, big.NewInt(1e9))

			// Make sure it doesn't already have 32 ETH in it
			requiredBalance := eth.EthToWei(32)
			if mpDetails.BeaconBalance.Cmp(requiredBalance) >= 0 {
				mpDetails.HasFullBalance = true
			}
		}

		mpDetails.CanRescue = !(mpDetails.IsFinalized || mpDetails.InvalidElState || mpDetails.InvalidBeaconState || mpDetails.HasFullBalance)
	}

	data.Details = details
	return types.ResponseStatus_Success, nil
}
