package node

import (
	"fmt"
	"math"
	"math/big"
	"net/url"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/v2/minipool"
	"github.com/rocket-pool/rocketpool-go/v2/tokens"
	rptypes "github.com/rocket-pool/rocketpool-go/v2/types"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	rprewards "github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/rewards"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/state"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

const (
	nodeShareBatchSize int = 200
)

// ===============
// === Factory ===
// ===============

type nodeRewardsContextFactory struct {
	handler *NodeHandler
}

func (f *nodeRewardsContextFactory) Create(args url.Values) (*nodeRewardsContext, error) {
	c := &nodeRewardsContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *nodeRewardsContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*nodeRewardsContext, api.NodeRewardsData](
		router, "rewards", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeRewardsContext struct {
	handler *NodeHandler
}

func (c *nodeRewardsContext) PrepareData(data *api.NodeRewardsData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	cfg := sp.GetConfig()
	rp := sp.GetRocketPool()
	ec := sp.GetEthClient()
	bc := sp.GetBeaconClient()
	ctx := c.handler.ctx
	nodeAddress, _ := sp.GetWallet().GetAddress()

	// Requirements
	status, err := sp.RequireNodeRegistered(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	rpl, err := tokens.NewTokenRpl(rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting RPL token binding: %w", err)
	}
	pMgr, err := protocol.NewProtocolDaoManager(rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting pDAO manager binding: %w", err)
	}

	// Details that aren't in the state
	// TODO: add these to the state
	var percentages protocol.RplRewardsPercentages
	err = rp.Query(func(mc *batch.MultiCaller) error {
		eth.AddQueryablesToMulticall(mc,
			rpl.TotalSupply,
		)
		pMgr.GetRewardsPercentages(mc, &percentages)
		return nil
	}, nil)

	// This thing is so complex it's easier to just get the state snapshot and go from there
	stateMgr, err := state.NewNetworkStateManager(ctx, rp, cfg, ec, bc, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating network state manager: %w", err)
	}
	mpMgr, err := minipool.NewMinipoolManager(rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating minipool manager binding: %w", err)
	}
	state, totalEffectiveStake, err := stateMgr.GetHeadStateForNode(ctx, nodeAddress, true)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting network state for node %s: %w", nodeAddress.Hex(), err)
	}

	// Some basic details
	node := state.NodeDetailsByAddress[nodeAddress]
	data.Registered = node.Exists
	data.NodeRegistrationTime = time.Unix(int64(node.RegistrationTime.Uint64()), 0)
	data.LastCheckpoint = state.NetworkDetails.IntervalStart
	data.RewardsInterval = state.NetworkDetails.IntervalDuration
	data.EffectiveRplStake = eth.WeiToEth(node.EffectiveRPLStake)
	data.TotalRplStake = eth.WeiToEth(node.RplStake)
	data.Trusted = false
	for _, odaoDetails := range state.OracleDaoMemberDetails {
		if odaoDetails.Address == nodeAddress {
			data.Trusted = true
			break
		}
	}

	// Rewards claim status
	unclaimedRplRewards := big.NewInt(0)
	unclaimedEthRewards := big.NewInt(0)
	claimedRplRewards := big.NewInt(0)
	claimedEthRewards := big.NewInt(0)
	unclaimedODaoRplRewards := big.NewInt(0)
	claimedODaoRplRewards := big.NewInt(0)
	claimStatus, err := rprewards.GetClaimStatus(rp, nodeAddress, state.NetworkDetails.RewardIndex)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting rewards claim status for node %s: %w", nodeAddress.Hex(), err)
	}
	for _, claimed := range claimStatus.Claimed {
		intervalInfo, err := rprewards.GetIntervalInfo(rp, cfg, nodeAddress, claimed, nil)
		if err != nil {
			return types.ResponseStatus_Error, err
		}
		if !intervalInfo.TreeFileExists {
			return types.ResponseStatus_ResourceConflict, fmt.Errorf("error calculating lifetime node rewards: rewards file %s doesn't exist but interval %d was claimed", intervalInfo.TreeFilePath, claimed)
		}
		claimedRplRewards.Add(claimedRplRewards, &intervalInfo.CollateralRplAmount.Int)
		claimedODaoRplRewards.Add(claimedODaoRplRewards, &intervalInfo.ODaoRplAmount.Int)
		claimedEthRewards.Add(claimedEthRewards, &intervalInfo.SmoothingPoolEthAmount.Int)
	}
	for _, unclaimed := range claimStatus.Unclaimed {
		intervalInfo, err := rprewards.GetIntervalInfo(rp, cfg, nodeAddress, unclaimed, nil)
		if err != nil {
			return types.ResponseStatus_Error, err
		}
		if !intervalInfo.TreeFileExists {
			return types.ResponseStatus_ResourceConflict, fmt.Errorf("error calculating lifetime node rewards: rewards file %s doesn't exist and interval %d is unclaimed", intervalInfo.TreeFilePath, unclaimed)
		}
		if intervalInfo.NodeExists {
			unclaimedRplRewards.Add(unclaimedRplRewards, &intervalInfo.CollateralRplAmount.Int)
			unclaimedODaoRplRewards.Add(unclaimedODaoRplRewards, &intervalInfo.ODaoRplAmount.Int)
			unclaimedEthRewards.Add(unclaimedEthRewards, &intervalInfo.SmoothingPoolEthAmount.Int)
		}
	}
	data.CumulativeRplRewards = eth.WeiToEth(claimedRplRewards)
	data.UnclaimedRplRewards = eth.WeiToEth(unclaimedRplRewards)
	data.CumulativeTrustedRplRewards = eth.WeiToEth(claimedODaoRplRewards)
	data.UnclaimedTrustedRplRewards = eth.WeiToEth(unclaimedODaoRplRewards)
	data.CumulativeEthRewards = eth.WeiToEth(claimedEthRewards)
	data.UnclaimedEthRewards = eth.WeiToEth(unclaimedEthRewards)

	// Calculate the estimated rewards
	rewardsIntervalDays := data.RewardsInterval.Seconds() / (60 * 60 * 24)
	inflationPerDay := eth.WeiToEth(state.NetworkDetails.RPLInflationIntervalRate)
	totalRplAtNextCheckpoint := (math.Pow(inflationPerDay, float64(rewardsIntervalDays)) - 1) * eth.WeiToEth(rpl.TotalSupply.Get())
	if totalRplAtNextCheckpoint < 0 {
		totalRplAtNextCheckpoint = 0
	}
	if totalEffectiveStake.Cmp(big.NewInt(0)) == 1 {
		data.EstimatedRewards = data.EffectiveRplStake / eth.WeiToEth(totalEffectiveStake) * totalRplAtNextCheckpoint * eth.WeiToEth(percentages.NodePercentage)
	}
	if data.Trusted {
		data.EstimatedTrustedRplRewards = totalRplAtNextCheckpoint * eth.WeiToEth(percentages.OdaoPercentage) / float64(len(state.OracleDaoMemberDetails))
	}

	// Get the Beacon rewards
	epoch := state.BeaconSlotNumber / state.BeaconConfig.SlotsPerEpoch
	var totalDepositBalance float64
	var totalNodeShare float64
	mps := state.MinipoolDetailsByNode[nodeAddress]
	mpsToCalcNodeShareFor := []*minipool.MinipoolCommon{}
	beaconBalances := []*big.Int{}
	for _, mpd := range mps {
		mp, err := mpMgr.NewMinipoolFromVersion(mpd.MinipoolAddress, mpd.Version)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error creating binding for minipool %s: %w", mpd.MinipoolAddress.Hex(), err)
		}
		validator := state.ValidatorDetails[mpd.Pubkey]
		blockBalance := eth.GweiToWei(float64(validator.Balance))

		// Data
		status := mpd.Status
		nodeDepositBalance := mpd.NodeDepositBalance
		finalized := mpd.Finalised

		// Deal with pools that haven't received deposits yet so their balance is still 0
		if nodeDepositBalance == nil {
			nodeDepositBalance = big.NewInt(0)
		}

		// Ignore finalized minipools
		if finalized {
			continue
		}

		// Use node deposit balance if initialized or prelaunch
		if status == rptypes.MinipoolStatus_Initialized || status == rptypes.MinipoolStatus_Prelaunch {
			totalDepositBalance += eth.WeiToEth(nodeDepositBalance)
			totalNodeShare += eth.WeiToEth(nodeDepositBalance)
			continue
		}

		// Use node deposit balance if validator not yet active on beacon chain at block
		if !validator.Exists || validator.ActivationEpoch >= epoch {
			totalDepositBalance += eth.WeiToEth(nodeDepositBalance)
			totalNodeShare += eth.WeiToEth(nodeDepositBalance)
			continue
		}

		// Add this to the list of MPs to get the node share for
		totalDepositBalance += eth.WeiToEth(nodeDepositBalance)
		mpsToCalcNodeShareFor = append(mpsToCalcNodeShareFor, mp.Common())
		beaconBalances = append(beaconBalances, blockBalance)
	}

	// Get node shares in batches
	nodeShares := make([]*big.Int, len(mpsToCalcNodeShareFor))
	err = rp.BatchQuery(len(mpsToCalcNodeShareFor), nodeShareBatchSize, func(mc *batch.MultiCaller, i int) error {
		mpsToCalcNodeShareFor[i].CalculateNodeShare(mc, &nodeShares[i], beaconBalances[i])
		return nil
	}, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error calculating node shares of beacon balances: %w", err)
	}

	// Sum up the node shares
	for _, nodeShare := range nodeShares {
		totalNodeShare += eth.WeiToEth(nodeShare)
	}
	data.BeaconRewards = totalNodeShare - totalDepositBalance
	return types.ResponseStatus_Success, nil
}
