package node

import (
	"fmt"
	"math/big"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/v2/network"
	"github.com/rocket-pool/rocketpool-go/v2/node"
	"github.com/rocket-pool/rocketpool-go/v2/rewards"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/collateral"
	rprewards "github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/rewards"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type nodeGetRewardsInfoContextFactory struct {
	handler *NodeHandler
}

func (f *nodeGetRewardsInfoContextFactory) Create(args url.Values) (*nodeGetRewardsInfoContext, error) {
	c := &nodeGetRewardsInfoContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *nodeGetRewardsInfoContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*nodeGetRewardsInfoContext, api.NodeGetRewardsInfoData](
		router, "get-rewards-info", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeGetRewardsInfoContext struct {
	handler *NodeHandler
	cfg     *config.SmartNodeConfig
	rp      *rocketpool.RocketPool

	node        *node.Node
	networkMgr  *network.NetworkManager
	pSettings   *protocol.ProtocolDaoSettings
	rewardsPool *rewards.RewardsPool
}

func (c *nodeGetRewardsInfoContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.cfg = sp.GetConfig()
	c.rp = sp.GetRocketPool()
	nodeAddress, _ := sp.GetWallet().GetAddress()

	// Requirements
	status, err := sp.RequireNodeRegistered(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	c.node, err = node.NewNode(c.rp, nodeAddress)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating node %s binding: %w", nodeAddress.Hex(), err)
	}
	c.networkMgr, err = network.NewNetworkManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating network manager binding: %w", err)
	}
	pMgr, err := protocol.NewProtocolDaoManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating pDAO settings binding: %w", err)
	}
	c.pSettings = pMgr.Settings
	c.rewardsPool, err = rewards.NewRewardsPool(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating rewards pool binding: %w", err)
	}
	return types.ResponseStatus_Success, nil
}

func (c *nodeGetRewardsInfoContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.node.ActiveMinipoolCount,
		c.node.RplStake,
		c.node.MinimumRplStake,
		c.node.EffectiveRplStake,
		c.networkMgr.RplPrice,
		c.pSettings.Node.MinimumPerMinipoolStake,
		c.rewardsPool.RewardIndex,
	)
}

func (c *nodeGetRewardsInfoContext) PrepareData(data *api.NodeGetRewardsInfoData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	// Basic details
	data.RplPrice = c.networkMgr.RplPrice.Raw()
	data.RplStake = c.node.RplStake.Get()
	data.MinimumRplStake = c.node.MinimumRplStake.Get()
	data.EffectiveRplStake = c.node.EffectiveRplStake.Get()

	// Get the claimed and unclaimed intervals
	claimStatus, err := rprewards.GetClaimStatus(c.rp, c.node.Address, c.rewardsPool.RewardIndex.Formatted())
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting rewards claim status: %w", err)
	}
	data.ClaimedIntervals = claimStatus.Claimed

	// Get the info for each unclaimed interval
	for _, unclaimedInterval := range claimStatus.Unclaimed {
		intervalInfo, err := rprewards.GetIntervalInfo(c.rp, c.cfg, c.node.Address, unclaimedInterval, nil)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting interval %d info: %w", unclaimedInterval, err)
		}
		if !intervalInfo.TreeFileExists || !intervalInfo.MerkleRootValid {
			data.InvalidIntervals = append(data.InvalidIntervals, intervalInfo)
			continue
		}
		if intervalInfo.NodeExists {
			data.UnclaimedIntervals = append(data.UnclaimedIntervals, intervalInfo)
		}
	}

	// Get the number of active (non-finalized) minipools
	data.ActiveMinipools = c.node.ActiveMinipoolCount.Formatted()
	if data.ActiveMinipools > 0 {
		collateral, err := collateral.CheckCollateral(c.rp, c.node.Address, nil)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting node collateral: %w", err)
		}
		data.EthMatched = collateral.EthMatched
		data.EthMatchedLimit = collateral.EthMatchedLimit
		data.PendingMatchAmount = collateral.PendingMatchAmount

		// Calculate the *real* minimum, including the pending bond reductions
		minStakeFraction := c.pSettings.Node.MinimumPerMinipoolStake.Raw()
		trueMinimumStake := big.NewInt(0).Add(data.EthMatched, data.PendingMatchAmount)
		trueMinimumStake.Mul(trueMinimumStake, minStakeFraction)
		trueMinimumStake.Div(trueMinimumStake, data.RplPrice)

		data.MinimumRplStake = trueMinimumStake

		if data.EffectiveRplStake.Cmp(trueMinimumStake) < 0 {
			data.EffectiveRplStake.SetUint64(0)
		}

		data.BondedCollateralRatio = eth.WeiToEth(data.RplPrice) * eth.WeiToEth(data.RplStake) / (float64(data.ActiveMinipools)*32.0 - eth.WeiToEth(data.EthMatched) - eth.WeiToEth(data.PendingMatchAmount))
		data.BorrowedCollateralRatio = eth.WeiToEth(data.RplPrice) * eth.WeiToEth(data.RplStake) / (eth.WeiToEth(data.EthMatched) + eth.WeiToEth(data.PendingMatchAmount))
	} else {
		data.BorrowedCollateralRatio = -1
	}

	return types.ResponseStatus_Success, nil
}
