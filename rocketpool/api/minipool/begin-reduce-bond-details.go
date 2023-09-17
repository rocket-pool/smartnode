package minipool

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/settings"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/rocketpool/common/beacon"
	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	sharedtypes "github.com/rocket-pool/smartnode/shared/types"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type minipoolBeginReduceBondDetailsContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolBeginReduceBondDetailsContextFactory) Create(vars map[string]string) (*minipoolBeginReduceBondDetailsContext, error) {
	c := &minipoolBeginReduceBondDetailsContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *minipoolBeginReduceBondDetailsContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterMinipoolRoute[*minipoolBeginReduceBondDetailsContext, api.MinipoolBeginReduceBondDetailsData](
		router, "begin-reduce-bond/details", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolBeginReduceBondDetailsContext struct {
	handler *MinipoolHandler
	rp      *rocketpool.RocketPool
	bc      beacon.Client

	newBondAmountWei *big.Int
	pSettings        *settings.ProtocolDaoSettings
	oSettings        *settings.OracleDaoSettings
}

func (c *minipoolBeginReduceBondDetailsContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.bc = sp.GetBeaconClient()

	// Requirements
	err := errors.Join(
		sp.RequireNodeRegistered(),
	)
	if err != nil {
		return err
	}

	c.pSettings, err = settings.NewProtocolDaoSettings(c.rp)
	if err != nil {
		return fmt.Errorf("error creating pDAO settings binding: %w", err)
	}
	c.oSettings, err = settings.NewOracleDaoSettings(c.rp)
	if err != nil {
		return fmt.Errorf("error creating oDAO settings binding: %w", err)
	}
	return nil
}

func (c *minipoolBeginReduceBondDetailsContext) GetState(node *node.Node, mc *batch.MultiCaller) {
	c.pSettings.GetBondReductionEnabled(mc)
	c.oSettings.GetBondReductionWindowStart(mc)
	c.oSettings.GetBondReductionWindowLength(mc)
}

func (c *minipoolBeginReduceBondDetailsContext) CheckState(node *node.Node, response *api.MinipoolBeginReduceBondDetailsData) bool {
	response.BondReductionDisabled = !c.pSettings.Details.Minipool.IsBondReductionEnabled
	return !response.BondReductionDisabled
}

func (c *minipoolBeginReduceBondDetailsContext) GetMinipoolDetails(mc *batch.MultiCaller, mp minipool.Minipool, index int) {
	mpv3, success := minipool.GetMinipoolAsV3(mp)
	if success {
		mpv3.GetNodeDepositBalance(mc)
		mpv3.GetFinalised(mc)
		mpv3.GetStatus(mc)
		mpv3.GetPubkey(mc)
		mpv3.GetReduceBondTime(mc)
	}
}

func (c *minipoolBeginReduceBondDetailsContext) PrepareData(addresses []common.Address, mps []minipool.Minipool, response *api.MinipoolBeginReduceBondDetailsData) error {
	// Get the latest block header
	header, err := c.rp.Client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("error getting latest block header: %w", err)
	}
	currentTime := time.Unix(int64(header.Time), 0)

	// Get the bond reduction details
	pubkeys := []types.ValidatorPubkey{}
	detailsMap := map[types.ValidatorPubkey]int{}
	details := make([]api.MinipoolBeginReduceBondDetails, len(addresses))
	for i, mp := range mps {
		mpCommon := mp.GetMinipoolCommon()
		mpDetails := api.MinipoolBeginReduceBondDetails{
			Address: mpCommon.Details.Address,
		}

		mpv3, success := minipool.GetMinipoolAsV3(mp)
		if !success {
			mpDetails.MinipoolVersionTooLow = true
		} else if mpCommon.Details.Status.Formatted() != types.Staking || mpCommon.Details.IsFinalised {
			mpDetails.InvalidElState = true
		} else {
			reductionStart := mpv3.Details.ReduceBondTime.Formatted()
			timeSinceBondReductionStart := currentTime.Sub(reductionStart)
			windowStart := c.oSettings.Details.Minipools.BondReductionWindowStart.Formatted()
			windowEnd := windowStart + c.oSettings.Details.Minipools.BondReductionWindowLength.Formatted()

			if timeSinceBondReductionStart < windowEnd {
				mpDetails.AlreadyInWindow = true
			} else {
				mpDetails.MatchRequest = big.NewInt(0).Sub(mpCommon.Details.NodeDepositBalance, c.newBondAmountWei)
				pubkeys = append(pubkeys, mpCommon.Details.Pubkey)
				detailsMap[mpCommon.Details.Pubkey] = i
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

		mpDetails.CanReduce = !(response.BondReductionDisabled || mpDetails.MinipoolVersionTooLow || mpDetails.AlreadyInWindow || mpDetails.BalanceTooLow || mpDetails.InvalidBeaconState)
	}

	response.Details = details
	return nil
}
