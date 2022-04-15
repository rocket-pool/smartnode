package rewards

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

// Get the index of the active rewards period
func GetRewardIndex(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketRewardsPool, err := getRocketRewardsPool(rp)
	if err != nil {
		return nil, err
	}
	index := new(*big.Int)
	if err := rocketRewardsPool.Call(opts, index, "getRewardIndex"); err != nil {
		return nil, fmt.Errorf("Could not get current reward index: %w", err)
	}
	return *index, nil
}

// Get the timestamp that the current rewards interval started
func GetClaimIntervalTimeStart(rp *rocketpool.RocketPool, opts *bind.CallOpts) (time.Time, error) {
	rocketRewardsPool, err := getRocketRewardsPool(rp)
	if err != nil {
		return time.Time{}, err
	}
	unixTime := new(*big.Int)
	if err := rocketRewardsPool.Call(opts, unixTime, "getClaimIntervalTimeStart"); err != nil {
		return time.Time{}, fmt.Errorf("Could not get claim interval time start: %w", err)
	}
	return time.Unix(int64((*unixTime).Uint64()), 0), nil
}

// Get the number of seconds in a claim interval
func GetClaimIntervalTime(rp *rocketpool.RocketPool, opts *bind.CallOpts) (time.Duration, error) {
	rocketRewardsPool, err := getRocketRewardsPool(rp)
	if err != nil {
		return 0, err
	}
	unixTime := new(*big.Int)
	if err := rocketRewardsPool.Call(opts, unixTime, "getClaimIntervalTime"); err != nil {
		return 0, fmt.Errorf("Could not get claim interval time: %w", err)
	}
	return time.Duration((*unixTime).Int64()) * time.Second, nil
}

// Get the percent of checkpoint rewards that goes to node operators
func GetNodeOperatorRewardsPercent(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketRewardsPool, err := getRocketRewardsPool(rp)
	if err != nil {
		return nil, err
	}
	perc := new(*big.Int)
	if err := rocketRewardsPool.Call(opts, perc, "getClaimingContractPerc", "rocketClaimNode"); err != nil {
		return nil, fmt.Errorf("Could not get node operator rewards percent: %w", err)
	}
	return *perc, nil
}

// Get the percent of checkpoint rewards that goes to ODAO members
func GetTrustedNodeOperatorRewardsPercent(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketRewardsPool, err := getRocketRewardsPool(rp)
	if err != nil {
		return nil, err
	}
	perc := new(*big.Int)
	if err := rocketRewardsPool.Call(opts, perc, "getClaimingContractPerc", "rocketClaimTrustedNode"); err != nil {
		return nil, fmt.Errorf("Could not get trusted node operator rewards percent: %w", err)
	}
	return *perc, nil
}

// Get the amount of RPL rewards that will be provided to node operators
func GetPendingRPLRewards(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketRewardsPool, err := getRocketRewardsPool(rp)
	if err != nil {
		return nil, err
	}
	rewards := new(*big.Int)
	if err := rocketRewardsPool.Call(opts, rewards, "getPendingRPLRewards"); err != nil {
		return nil, fmt.Errorf("Could not get pending RPL rewards: %w", err)
	}
	return *rewards, nil
}

// Get the amount of ETH rewards that will be provided to node operators
func GetPendingETHRewards(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketRewardsPool, err := getRocketRewardsPool(rp)
	if err != nil {
		return nil, err
	}
	rewards := new(*big.Int)
	if err := rocketRewardsPool.Call(opts, rewards, "getPendingETHRewards"); err != nil {
		return nil, fmt.Errorf("Could not get pending ETH rewards: %w", err)
	}
	return *rewards, nil
}

// Submit a Merkle Tree-based snapshot for a rewards interval
func SubmitRewardSnapshot(rp *rocketpool.RocketPool, index *big.Int, block *big.Int, rewardsPerNetworkRPL []*big.Int, rewardsPerNetworkETH []*big.Int, merkleRoot []byte, merkleTreeCID string, opts *bind.TransactOpts) (common.Hash, error) {
	rocketRewardsPool, err := getRocketRewardsPool(rp)
	if err != nil {
		return common.Hash{}, err
	}
	hash, err := rocketRewardsPool.Transact(opts, "submitRewardSnapshot", index, rewardsPerNetworkRPL, rewardsPerNetworkETH, merkleRoot, merkleTreeCID)
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not submit rewards snapshot: %w", err)
	}
	return hash, nil
}

// Get the start block of the provided rewards interval
func GetStartBlockOfInterval(rp *rocketpool.RocketPool, index uint64, opts *bind.CallOpts) (uint64, error) {
	// 0 is a special case because the start time is based on the legacy rewards system
	if index == 0 {
		// Get the start time of the final legacy rewards interval
		lastLegacyIntervalStartHash := crypto.Keccak256Hash([]byte("rewards.pool.claim.interval.time.last"))
		rp.RocketStorage.GetUint(nil, lastLegacyIntervalStartHash)
	} else {
		// Find it based on the events
	}

	return 0, nil
}

// Get contracts
var rocketRewardsPoolLock sync.Mutex

func getRocketRewardsPool(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
	rocketRewardsPoolLock.Lock()
	defer rocketRewardsPoolLock.Unlock()
	return rp.GetContract("rocketRewardsPool")
}
