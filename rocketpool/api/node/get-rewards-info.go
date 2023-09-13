package node

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/settings"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rprewards "github.com/rocket-pool/smartnode/shared/services/rewards"
	"github.com/rocket-pool/smartnode/shared/types/api"
	rputils "github.com/rocket-pool/smartnode/shared/utils/rp"
)

type nodeRewardsInfoHandler struct {
	networkPrices *network.NetworkPrices
	pSettings     *settings.ProtocolDaoSettings
	rewardsPool   *rewards.RewardsPool
}

func (h *nodeRewardsInfoHandler) CreateBindings(rp *rocketpool.RocketPool) error {
	var err error
	h.networkPrices, err = network.NewNetworkPrices(rp)
	if err != nil {
		return fmt.Errorf("error creating network prices binding: %w", err)
	}
	h.pSettings, err = settings.NewProtocolDaoSettings(rp)
	if err != nil {
		return fmt.Errorf("error creating pDAO settings binding: %w", err)
	}
	h.rewardsPool, err = rewards.NewRewardsPool(rp)
	if err != nil {
		return fmt.Errorf("error creating rewards pool binding: %w", err)
	}
	return nil
}

func (h *nodeRewardsInfoHandler) GetState(node *node.Node, mc *batch.MultiCaller) {
	node.GetActiveMinipoolCount(mc)
	node.GetRplStake(mc)
	node.GetMinimumRplStake(mc)
	node.GetMaximumRplStake(mc)
	node.GetEffectiveRplStake(mc)
	h.networkPrices.GetRplPrice(mc)
	h.pSettings.GetMinimumPerMinipoolStake(mc)
	h.pSettings.GetMaximumPerMinipoolStake(mc)
	h.rewardsPool.GetRewardIndex(mc)
}

func (h *nodeRewardsInfoHandler) PrepareResponse(rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, node *node.Node, opts *bind.TransactOpts, response *api.NodeGetRewardsInfoResponse) error {
	// Basic details
	response.RplPrice = h.networkPrices.Details.RplPrice.RawValue
	response.RplStake = node.Details.RplStake
	response.MinimumRplStake = node.Details.MinimumRplStake
	response.MaximumRplStake = node.Details.MaximumRplStake
	response.EffectiveRplStake = node.Details.EffectiveRplStake

	// Get the claimed and unclaimed intervals
	claimStatus, err := rprewards.GetClaimStatus(rp, node.Details.Address, h.rewardsPool.Details.RewardIndex.Formatted())
	if err != nil {
		return fmt.Errorf("error getting rewards claim status: %w", err)
	}
	response.ClaimedIntervals = claimStatus.Claimed

	// Get the info for each unclaimed interval
	for _, unclaimedInterval := range claimStatus.Unclaimed {
		intervalInfo, err := rprewards.GetIntervalInfo(rp, cfg, node.Details.Address, unclaimedInterval, nil)
		if err != nil {
			return fmt.Errorf("error getting interval %d info: %w", unclaimedInterval, err)
		}
		if !intervalInfo.TreeFileExists || !intervalInfo.MerkleRootValid {
			response.InvalidIntervals = append(response.InvalidIntervals, intervalInfo)
			continue
		}
		if intervalInfo.NodeExists {
			response.UnclaimedIntervals = append(response.UnclaimedIntervals, intervalInfo)
		}
	}

	// Get the number of active (non-finalized) minipools
	response.ActiveMinipools = node.Details.ActiveMinipoolCount.Formatted()
	if response.ActiveMinipools > 0 {
		collateral, err := rputils.CheckCollateral(rp, node.Details.Address, nil)
		if err != nil {
			return fmt.Errorf("error getting node collateral: %w", err)
		}
		response.EthMatched = collateral.EthMatched
		response.EthMatchedLimit = collateral.EthMatchedLimit
		response.PendingMatchAmount = collateral.PendingMatchAmount

		// Calculate the *real* minimum, including the pending bond reductions
		minStakeFraction := h.pSettings.Details.Node.MinimumPerMinipoolStake.RawValue
		maxStakeFraction := h.pSettings.Details.Node.MaximumPerMinipoolStake.RawValue
		trueMinimumStake := big.NewInt(0).Add(response.EthMatched, response.PendingMatchAmount)
		trueMinimumStake.Mul(trueMinimumStake, minStakeFraction)
		trueMinimumStake.Div(trueMinimumStake, response.RplPrice)

		// Calculate the *real* maximum, including the pending bond reductions
		trueMaximumStake := eth.EthToWei(32)
		trueMaximumStake.Mul(trueMaximumStake, big.NewInt(int64(response.ActiveMinipools)))
		trueMaximumStake.Sub(trueMaximumStake, response.EthMatched)
		trueMaximumStake.Sub(trueMaximumStake, response.PendingMatchAmount) // (32 * activeMinipools - ethMatched - pendingMatch)
		trueMaximumStake.Mul(trueMaximumStake, maxStakeFraction)
		trueMaximumStake.Div(trueMaximumStake, response.RplPrice)

		response.MinimumRplStake = trueMinimumStake
		response.MaximumRplStake = trueMaximumStake

		if response.EffectiveRplStake.Cmp(trueMinimumStake) < 0 {
			response.EffectiveRplStake.SetUint64(0)
		} else if response.EffectiveRplStake.Cmp(trueMaximumStake) > 0 {
			response.EffectiveRplStake.Set(trueMaximumStake)
		}

		response.BondedCollateralRatio = eth.WeiToEth(response.RplPrice) * eth.WeiToEth(response.RplStake) / (float64(response.ActiveMinipools)*32.0 - eth.WeiToEth(response.EthMatched) - eth.WeiToEth(response.PendingMatchAmount))
		response.BorrowedCollateralRatio = eth.WeiToEth(response.RplPrice) * eth.WeiToEth(response.RplStake) / (eth.WeiToEth(response.EthMatched) + eth.WeiToEth(response.PendingMatchAmount))
	} else {
		response.BorrowedCollateralRatio = -1
	}

	return nil
}
