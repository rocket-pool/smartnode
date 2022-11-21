package rewards

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	rewards_v150rc1 "github.com/rocket-pool/rocketpool-go/legacy/v1.5.0-rc1/rewards"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services/config"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

// This retrieves the rewards snapshot event from a set of contracts, upgrading it to the latest struct version
func GetUpgradedRewardSnapshotEvent(cfg *config.RocketPoolConfig, rp *rocketpool.RocketPool, index uint64, intervalSize *big.Int, startBlock *big.Int, endBlock *big.Int) (rewards.RewardsEvent, error) {

	// Get the version map
	versionMap := cfg.Smartnode.GetPreviousRewardsPoolAddresses()

	// Check old versions
	for version, addresses := range versionMap {
		switch version {
		case "v1.5.0-rc1":
			found, oldRewardsEvent, err := rewards_v150rc1.GetRewardSnapshotEventWithUpgrades(rp, index, intervalSize, startBlock, endBlock, addresses, nil)
			if err != nil {
				return rewards.RewardsEvent{}, fmt.Errorf("error checking %s contracts for rewards event %d: %w", version, index, err)
			}
			if found {
				return update_v150rc1_to_v150(oldRewardsEvent), nil
			}
		}
	}

	// Check the current contract
	return rewards.GetRewardSnapshotEvent(rp, index, intervalSize, startBlock, endBlock, nil)

}

// Upgrade a rewards event from v1.5.0-RC1 to v1.5.0
func update_v150rc1_to_v150(oldEvent rewards_v150rc1.RewardsEvent) rewards.RewardsEvent {
	newEvent := rewards.RewardsEvent{
		Index:             oldEvent.Index,
		ExecutionBlock:    oldEvent.ExecutionBlock,
		ConsensusBlock:    oldEvent.ConsensusBlock,
		MerkleRoot:        oldEvent.MerkleRoot,
		MerkleTreeCID:     oldEvent.MerkleTreeCID,
		IntervalsPassed:   oldEvent.IntervalsPassed,
		TreasuryRPL:       oldEvent.TreasuryRPL,
		TrustedNodeRPL:    oldEvent.TrustedNodeRPL,
		NodeRPL:           oldEvent.NodeRPL,
		NodeETH:           oldEvent.NodeETH,
		UserETH:           big.NewInt(0),
		IntervalStartTime: oldEvent.IntervalStartTime,
		IntervalEndTime:   oldEvent.IntervalEndTime,
		SubmissionTime:    oldEvent.SubmissionTime,
	}

	return newEvent
}

// TODO: temp until rocketpool-go supports RocketStorage contract address lookups per block
func GetClaimIntervalTime(cfg *config.RocketPoolConfig, index uint64, rp *rocketpool.RocketPool, opts *bind.CallOpts) (time.Duration, error) {
	switch cfg.Smartnode.Network.Value.(cfgtypes.Network) {
	case cfgtypes.Network_Prater:
		if index < 2 {
			contractAddress := cfg.Smartnode.GetPreviousRewardsPoolAddresses()["v1.5.0-rc1"][0]
			return rewards_v150rc1.GetClaimIntervalTime(rp, opts, &contractAddress)
		}
	}

	return rewards.GetClaimIntervalTime(rp, opts)
}

// TODO: temp until rocketpool-go supports RocketStorage contract address lookups per block
func GetNodeOperatorRewardsPercent(cfg *config.RocketPoolConfig, index uint64, rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	switch cfg.Smartnode.Network.Value.(cfgtypes.Network) {
	case cfgtypes.Network_Prater:
		if index < 2 {
			contractAddress := cfg.Smartnode.GetPreviousRewardsPoolAddresses()["v1.5.0-rc1"][0]
			return rewards_v150rc1.GetNodeOperatorRewardsPercent(rp, opts, &contractAddress)
		}
	}

	return rewards.GetNodeOperatorRewardsPercent(rp, opts)
}

// TODO: temp until rocketpool-go supports RocketStorage contract address lookups per block
func GetTrustedNodeOperatorRewardsPercent(cfg *config.RocketPoolConfig, index uint64, rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	switch cfg.Smartnode.Network.Value.(cfgtypes.Network) {
	case cfgtypes.Network_Prater:
		if index < 2 {
			contractAddress := cfg.Smartnode.GetPreviousRewardsPoolAddresses()["v1.5.0-rc1"][0]
			return rewards_v150rc1.GetTrustedNodeOperatorRewardsPercent(rp, opts, &contractAddress)
		}
	}

	return rewards.GetTrustedNodeOperatorRewardsPercent(rp, opts)
}

// TODO: temp until rocketpool-go supports RocketStorage contract address lookups per block
func GetProtocolDaoRewardsPercent(cfg *config.RocketPoolConfig, index uint64, rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	switch cfg.Smartnode.Network.Value.(cfgtypes.Network) {
	case cfgtypes.Network_Prater:
		if index < 2 {
			contractAddress := cfg.Smartnode.GetPreviousRewardsPoolAddresses()["v1.5.0-rc1"][0]
			return rewards_v150rc1.GetProtocolDaoRewardsPercent(rp, opts, &contractAddress)
		}
	}

	return rewards.GetProtocolDaoRewardsPercent(rp, opts)
}

// TODO: temp until rocketpool-go supports RocketStorage contract address lookups per block
func GetPendingRPLRewards(cfg *config.RocketPoolConfig, index uint64, rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	switch cfg.Smartnode.Network.Value.(cfgtypes.Network) {
	case cfgtypes.Network_Prater:
		if index < 2 {
			contractAddress := cfg.Smartnode.GetPreviousRewardsPoolAddresses()["v1.5.0-rc1"][0]
			return rewards_v150rc1.GetPendingRPLRewards(rp, opts, &contractAddress)
		}
	}

	return rewards.GetPendingRPLRewards(rp, opts)
}
