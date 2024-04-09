package minipool

import (
	"fmt"
	"net/url"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/dao/oracle"
	"github.com/rocket-pool/rocketpool-go/v2/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/v2/minipool"
	"github.com/rocket-pool/rocketpool-go/v2/node"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/v2/types"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type minipoolReduceBondDetailsContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolReduceBondDetailsContextFactory) Create(args url.Values) (*minipoolReduceBondDetailsContext, error) {
	c := &minipoolReduceBondDetailsContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *minipoolReduceBondDetailsContextFactory) RegisterRoute(router *mux.Router) {
	RegisterMinipoolRoute[*minipoolReduceBondDetailsContext, api.MinipoolReduceBondDetailsData](
		router, "reduce-bond/details", f, f.handler.ctx, f.handler.logger, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolReduceBondDetailsContext struct {
	handler *MinipoolHandler
	rp      *rocketpool.RocketPool
	bc      beacon.IBeaconClient

	node      *node.Node
	pSettings *protocol.ProtocolDaoSettings
	oSettings *oracle.OracleDaoSettings
}

func (c *minipoolReduceBondDetailsContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.bc = sp.GetBeaconClient()
	nodeAddress, _ := sp.GetWallet().GetAddress()

	// Requirements
	err := sp.RequireBeaconClientSynced(c.handler.ctx)
	if err != nil {
		return types.ResponseStatus_ClientsNotSynced, err
	}

	// Bindings
	c.node, err = node.NewNode(c.rp, nodeAddress)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating node binding: %w", err)
	}
	pMgr, err := protocol.NewProtocolDaoManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating pDAO manager binding: %w", err)
	}
	c.pSettings = pMgr.Settings
	oMgr, err := oracle.NewOracleDaoManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating oDAO manager binding: %w", err)
	}
	c.oSettings = oMgr.Settings
	return types.ResponseStatus_Success, nil
}

func (c *minipoolReduceBondDetailsContext) GetState(node *node.Node, mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.node.IsFeeDistributorInitialized,
		c.pSettings.Minipool.IsBondReductionEnabled,
		c.oSettings.Minipool.BondReductionWindowStart,
		c.oSettings.Minipool.BondReductionWindowLength,
	)
}

func (c *minipoolReduceBondDetailsContext) CheckState(node *node.Node, data *api.MinipoolReduceBondDetailsData) bool {
	data.BondReductionDisabled = !c.pSettings.Minipool.IsBondReductionEnabled.Get()
	return !data.BondReductionDisabled
}

func (c *minipoolReduceBondDetailsContext) GetMinipoolDetails(mc *batch.MultiCaller, mp minipool.IMinipool, index int) {
	mpCommon := mp.Common()
	eth.AddQueryablesToMulticall(mc,
		mpCommon.NodeDepositBalance,
		mpCommon.NodeFee,
	)
	mpv3, success := minipool.GetMinipoolAsV3(mp)
	if success {
		eth.AddQueryablesToMulticall(mc,
			mpv3.IsFinalised,
			mpv3.Status,
			mpv3.Pubkey,
			mpv3.ReduceBondTime,
			mpv3.IsBondReduceCancelled,
		)
	}
}

func (c *minipoolReduceBondDetailsContext) PrepareData(addresses []common.Address, mps []minipool.IMinipool, data *api.MinipoolReduceBondDetailsData) (types.ResponseStatus, error) {
	// General vars
	data.IsFeeDistributorInitialized = c.node.IsFeeDistributorInitialized.Get()
	data.BondReductionWindowStart = c.oSettings.Minipool.BondReductionWindowStart.Formatted()
	data.BondReductionWindowLength = c.oSettings.Minipool.BondReductionWindowLength.Formatted()

	// Get the latest block header
	ctx := c.handler.ctx
	header, err := c.rp.Client.HeaderByNumber(ctx, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting latest block header: %w", err)
	}
	currentTime := time.Unix(int64(header.Time), 0)

	// Get the bond reduction details
	pubkeys := []beacon.ValidatorPubkey{}
	detailsMap := map[beacon.ValidatorPubkey]int{}
	details := make([]api.MinipoolReduceBondDetails, len(addresses))
	for i, mp := range mps {
		details[i] = c.getMinipoolDetails(mp, currentTime)
		pubkey := mp.Common().Pubkey.Get()
		pubkeys = append(pubkeys, pubkey)
		detailsMap[pubkey] = i
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
		mpDetails.Balance = beaconStatus.Balance
		mpDetails.BeaconState = beaconStatus.Status

		// Check the beacon state
		mpDetails.InvalidBeaconState = !(mpDetails.BeaconState == beacon.ValidatorState_PendingInitialized ||
			mpDetails.BeaconState == beacon.ValidatorState_PendingQueued ||
			mpDetails.BeaconState == beacon.ValidatorState_ActiveOngoing)

		// Make sure the balance is high enough
		threshold := uint64(32000000000)
		mpDetails.BalanceTooLow = mpDetails.Balance < threshold

		mpDetails.CanReduce = !(data.BondReductionDisabled || mpDetails.MinipoolVersionTooLow || mpDetails.OutOfWindow || mpDetails.BalanceTooLow || mpDetails.InvalidBeaconState || mpDetails.AlreadyCancelled)
	}

	data.Details = details
	return types.ResponseStatus_Success, nil
}

func (c *minipoolReduceBondDetailsContext) getMinipoolDetails(mp minipool.IMinipool, currentTime time.Time) api.MinipoolReduceBondDetails {
	mpCommon := mp.Common()
	mpDetails := api.MinipoolReduceBondDetails{
		Address: mpCommon.Address,
	}
	mpDetails.NodeDepositBalance = mpCommon.NodeDepositBalance.Get()
	mpDetails.NodeFee = mpCommon.NodeFee.Raw()

	mpv3, success := minipool.GetMinipoolAsV3(mp)
	if !success {
		mpDetails.MinipoolVersionTooLow = true
		return mpDetails
	}
	if mpCommon.Status.Formatted() != rptypes.MinipoolStatus_Staking || mpCommon.IsFinalised.Get() {
		mpDetails.InvalidElState = true
		return mpDetails
	}
	if mpv3.IsBondReduceCancelled.Get() {
		mpDetails.AlreadyCancelled = true
		return mpDetails
	}

	reductionStart := mpv3.ReduceBondTime.Formatted()
	timeSinceBondReductionStart := currentTime.Sub(reductionStart)
	windowStart := c.oSettings.Minipool.BondReductionWindowStart.Formatted()
	windowEnd := windowStart + c.oSettings.Minipool.BondReductionWindowLength.Formatted()

	if timeSinceBondReductionStart < windowStart || timeSinceBondReductionStart > windowEnd {
		mpDetails.OutOfWindow = true
		return mpDetails
	}

	return mpDetails
}
