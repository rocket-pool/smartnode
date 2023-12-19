package minipool

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/dao/oracle"
	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/rocketpool/common/beacon"
	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	sharedtypes "github.com/rocket-pool/smartnode/shared/types"
	"github.com/rocket-pool/smartnode/shared/types/api"
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
	server.RegisterMinipoolRoute[*minipoolReduceBondDetailsContext, api.MinipoolReduceBondDetailsData](
		router, "reduce-bond/details", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolReduceBondDetailsContext struct {
	handler *MinipoolHandler
	rp      *rocketpool.RocketPool
	bc      beacon.Client

	pSettings *protocol.ProtocolDaoSettings
	oSettings *oracle.OracleDaoSettings
}

func (c *minipoolReduceBondDetailsContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.bc = sp.GetBeaconClient()

	// Requirements
	err := errors.Join(
		sp.RequireNodeRegistered(),
		sp.RequireBeaconClientSynced(),
	)
	if err != nil {
		return err
	}

	// Bindings
	pMgr, err := protocol.NewProtocolDaoManager(c.rp)
	if err != nil {
		return fmt.Errorf("error creating pDAO manager binding: %w", err)
	}
	c.pSettings = pMgr.Settings
	oMgr, err := oracle.NewOracleDaoManager(c.rp)
	if err != nil {
		return fmt.Errorf("error creating oDAO manager binding: %w", err)
	}
	c.oSettings = oMgr.Settings
	return nil
}

func (c *minipoolReduceBondDetailsContext) GetState(node *node.Node, mc *batch.MultiCaller) {
	core.AddQueryablesToMulticall(mc,
		c.pSettings.Minipool.IsBondReductionEnabled,
		c.oSettings.Minipool.BondReductionWindowStart,
		c.oSettings.Minipool.BondReductionWindowLength,
	)
}

func (c *minipoolReduceBondDetailsContext) CheckState(node *node.Node, response *api.MinipoolReduceBondDetailsData) bool {
	response.BondReductionDisabled = !c.pSettings.Minipool.IsBondReductionEnabled.Get()
	return !response.BondReductionDisabled
}

func (c *minipoolReduceBondDetailsContext) GetMinipoolDetails(mc *batch.MultiCaller, mp minipool.IMinipool, index int) {
	mpv3, success := minipool.GetMinipoolAsV3(mp)
	if success {
		core.AddQueryablesToMulticall(mc,
			mpv3.IsFinalised,
			mpv3.Status,
			mpv3.Pubkey,
			mpv3.ReduceBondTime,
		)
	}
}

func (c *minipoolReduceBondDetailsContext) PrepareData(addresses []common.Address, mps []minipool.IMinipool, data *api.MinipoolReduceBondDetailsData) error {
	// Get the latest block header
	header, err := c.rp.Client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("error getting latest block header: %w", err)
	}
	currentTime := time.Unix(int64(header.Time), 0)

	// Get the bond reduction details
	pubkeys := []types.ValidatorPubkey{}
	detailsMap := map[types.ValidatorPubkey]int{}
	details := make([]api.MinipoolReduceBondDetails, len(addresses))
	for i, mp := range mps {
		mpCommon := mp.Common()
		mpDetails := api.MinipoolReduceBondDetails{
			Address: mpCommon.Address,
		}

		mpv3, success := minipool.GetMinipoolAsV3(mp)
		if !success {
			mpDetails.MinipoolVersionTooLow = true
		} else if mpCommon.Status.Formatted() != types.MinipoolStatus_Staking || mpCommon.IsFinalised.Get() {
			mpDetails.InvalidElState = true
		} else {
			reductionStart := mpv3.ReduceBondTime.Formatted()
			timeSinceBondReductionStart := currentTime.Sub(reductionStart)
			windowStart := c.oSettings.Minipool.BondReductionWindowStart.Formatted()
			windowEnd := windowStart + c.oSettings.Minipool.BondReductionWindowLength.Formatted()

			if timeSinceBondReductionStart < windowStart || timeSinceBondReductionStart > windowEnd {
				mpDetails.OutOfWindow = true
			} else {
				pubkey := mpCommon.Pubkey.Get()
				pubkeys = append(pubkeys, pubkey)
				detailsMap[pubkey] = i
			}
		}

		details[i] = mpDetails
	}

	// Get the statuses on Beacon
	beaconStatuses, err := c.bc.GetValidatorStatuses(pubkeys, nil)
	if err != nil {
		return fmt.Errorf("error getting validator statuses on Beacon: %w", err)
	}

	// Do a complete viability check
	for pubkey, beaconStatus := range beaconStatuses {
		i := detailsMap[pubkey]
		mpDetails := &details[i]
		mpDetails.Balance = beaconStatus.Balance
		mpDetails.BeaconState = beaconStatus.Status

		// Check the beacon state
		mpDetails.InvalidBeaconState = !(mpDetails.BeaconState == sharedtypes.ValidatorState_PendingInitialized ||
			mpDetails.BeaconState == sharedtypes.ValidatorState_PendingQueued ||
			mpDetails.BeaconState == sharedtypes.ValidatorState_ActiveOngoing)

		// Make sure the balance is high enough
		threshold := uint64(32000000000)
		mpDetails.BalanceTooLow = mpDetails.Balance < threshold

		mpDetails.CanReduce = !(data.BondReductionDisabled || mpDetails.MinipoolVersionTooLow || mpDetails.OutOfWindow || mpDetails.BalanceTooLow || mpDetails.InvalidBeaconState)
	}

	data.Details = details
	return nil
}
