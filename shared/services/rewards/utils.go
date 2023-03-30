package rewards

import (
	"fmt"
	"math/big"

	rewards_v110rc1 "github.com/rocket-pool/rocketpool-go/legacy/v1.1.0-rc1/rewards"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services/config"
)

// This retrieves the rewards snapshot event from a set of contracts, upgrading it to the latest struct version
func GetUpgradedRewardSnapshotEvent(cfg *config.RocketPoolConfig, rp *rocketpool.RocketPool, index uint64, intervalSize *big.Int, startBlock *big.Int, endBlock *big.Int) (rewards.RewardsEvent, error) {

	// Get the version map
	versionMap := cfg.Smartnode.GetPreviousRewardsPoolAddresses()

	// Check old versions
	for version, addresses := range versionMap {
		switch version {
		case "v1.1.0-rc1":
			found, oldRewardsEvent, err := rewards_v110rc1.GetRewardSnapshotEventWithUpgrades(rp, index, intervalSize, startBlock, endBlock, addresses, nil)
			if err != nil {
				return rewards.RewardsEvent{}, fmt.Errorf("error checking %s contracts for rewards event %d: %w", version, index, err)
			}
			if found {
				return update_v110rc1_to_v110(oldRewardsEvent), nil
			}
		case "v1.1.0", "v1.2.0-rc1":
			found, event, err := rewards.GetRewardSnapshotEventWithUpgrades(rp, index, intervalSize, startBlock, endBlock, addresses, nil)
			if err != nil {
				return rewards.RewardsEvent{}, fmt.Errorf("error checking %s contracts for rewards event %d: %w", version, index, err)
			}
			if found {
				return event, nil
			}
		}
	}

	// Check the current contract
	return rewards.GetRewardSnapshotEvent(rp, index, intervalSize, startBlock, endBlock, nil)

}

// Upgrade a rewards event from v1.1.0-RC1 to v1.1.0
func update_v110rc1_to_v110(oldEvent rewards_v110rc1.RewardsEvent) rewards.RewardsEvent {
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
