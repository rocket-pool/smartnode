package rewards

import (
	"fmt"
	"math/big"
	"reflect"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
)

// Info for a rewards snapshot event
type RewardsEvent struct {
	Index             *big.Int
	ExecutionBlock    *big.Int
	ConsensusBlock    *big.Int
	MerkleRoot        common.Hash
	MerkleTreeCID     string
	IntervalsPassed   *big.Int
	TreasuryRPL       *big.Int
	TrustedNodeRPL    []*big.Int
	NodeRPL           []*big.Int
	NodeETH           []*big.Int
	UserETH           *big.Int
	IntervalStartTime time.Time
	IntervalEndTime   time.Time
	SubmissionTime    time.Time
}

// Struct for submitting the rewards for a checkpoint
type RewardSubmission struct {
	RewardIndex     *big.Int   `json:"rewardIndex"`
	ExecutionBlock  *big.Int   `json:"executionBlock"`
	ConsensusBlock  *big.Int   `json:"consensusBlock"`
	MerkleRoot      [32]byte   `json:"merkleRoot"`
	MerkleTreeCID   string     `json:"merkleTreeCID"`
	IntervalsPassed *big.Int   `json:"intervalsPassed"`
	TreasuryRPL     *big.Int   `json:"treasuryRPL"`
	TrustedNodeRPL  []*big.Int `json:"trustedNodeRPL"`
	NodeRPL         []*big.Int `json:"nodeRPL"`
	NodeETH         []*big.Int `json:"nodeETH"`
	UserETH         *big.Int   `json:"userETH"`
}

// Get the index of the active rewards period
func GetRewardIndex(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketRewardsPool, err := getRocketRewardsPool(rp, opts)
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
	rocketRewardsPool, err := getRocketRewardsPool(rp, opts)
	if err != nil {
		return time.Time{}, err
	}
	unixTime := new(*big.Int)
	if err := rocketRewardsPool.Call(opts, unixTime, "getClaimIntervalTimeStart"); err != nil {
		return time.Time{}, fmt.Errorf("Could not get claim interval time start: %w", err)
	}
	return time.Unix((*unixTime).Int64(), 0), nil
}

// Get the number of seconds in a claim interval
func GetClaimIntervalTime(rp *rocketpool.RocketPool, opts *bind.CallOpts) (time.Duration, error) {
	rocketRewardsPool, err := getRocketRewardsPool(rp, opts)
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
	rocketRewardsPool, err := getRocketRewardsPool(rp, opts)
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
	rocketRewardsPool, err := getRocketRewardsPool(rp, opts)
	if err != nil {
		return nil, err
	}
	perc := new(*big.Int)
	if err := rocketRewardsPool.Call(opts, perc, "getClaimingContractPerc", "rocketClaimTrustedNode"); err != nil {
		return nil, fmt.Errorf("Could not get trusted node operator rewards percent: %w", err)
	}
	return *perc, nil
}

// Get the percent of checkpoint rewards that goes to the PDAO
func GetProtocolDaoRewardsPercent(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketRewardsPool, err := getRocketRewardsPool(rp, opts)
	if err != nil {
		return nil, err
	}
	perc := new(*big.Int)
	if err := rocketRewardsPool.Call(opts, perc, "getClaimingContractPerc", "rocketClaimDAO"); err != nil {
		return nil, fmt.Errorf("Could not get protocol DAO rewards percent: %w", err)
	}
	return *perc, nil
}

// Get the amount of RPL rewards that will be provided to node operators
func GetPendingRPLRewards(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketRewardsPool, err := getRocketRewardsPool(rp, opts)
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
	rocketRewardsPool, err := getRocketRewardsPool(rp, opts)
	if err != nil {
		return nil, err
	}
	rewards := new(*big.Int)
	if err := rocketRewardsPool.Call(opts, rewards, "getPendingETHRewards"); err != nil {
		return nil, fmt.Errorf("Could not get pending ETH rewards: %w", err)
	}
	return *rewards, nil
}

// Estimate the gas for submiting a Merkle Tree-based snapshot for a rewards interval
func EstimateSubmitRewardSnapshotGas(rp *rocketpool.RocketPool, submission RewardSubmission, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketRewardsPool, err := getRocketRewardsPool(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketRewardsPool.GetTransactionGasInfo(opts, "submitRewardSnapshot", submission)
}

// Submit a Merkle Tree-based snapshot for a rewards interval
func SubmitRewardSnapshot(rp *rocketpool.RocketPool, submission RewardSubmission, opts *bind.TransactOpts) (common.Hash, error) {
	rocketRewardsPool, err := getRocketRewardsPool(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketRewardsPool.Transact(opts, "submitRewardSnapshot", submission)
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not submit rewards snapshot: %w", err)
	}
	return tx.Hash(), nil
}

// Get the event info for a rewards snapshot
func GetRewardSnapshotEvent(rp *rocketpool.RocketPool, index uint64, intervalSize *big.Int, startBlock *big.Int, endBlock *big.Int, opts *bind.CallOpts) (RewardsEvent, error) {
	// Get contracts
	rocketRewardsPool, err := getRocketRewardsPool(rp, opts)
	if err != nil {
		return RewardsEvent{}, err
	}

	// Construct a filter query for relevant logs
	indexBig := big.NewInt(0).SetUint64(index)
	indexBytes := [32]byte{}
	indexBig.FillBytes(indexBytes[:])
	addressFilter := []common.Address{*rocketRewardsPool.Address}
	topicFilter := [][]common.Hash{{rocketRewardsPool.ABI.Events["RewardSnapshot"].ID}, {indexBytes}}

	// Get the event logs
	logs, err := eth.GetLogs(rp, addressFilter, topicFilter, intervalSize, startBlock, endBlock, nil)
	if err != nil {
		return RewardsEvent{}, err
	}

	// Get the log info
	values := make(map[string]interface{})
	if len(logs) == 0 {
		return RewardsEvent{}, fmt.Errorf("reward snapshot for interval %d not found", index)
	}
	err = rocketRewardsPool.ABI.Events["RewardSnapshot"].Inputs.UnpackIntoMap(values, logs[0].Data)
	if err != nil {
		return RewardsEvent{}, err
	}

	// Get the decoded data
	submissionPrototype := RewardSubmission{}
	submissionType := reflect.TypeOf(submissionPrototype)
	submission := reflect.ValueOf(values["submission"]).Convert(submissionType).Interface().(RewardSubmission)
	eventIntervalStartTime := values["intervalStartTime"].(*big.Int)
	eventIntervalEndTime := values["intervalEndTime"].(*big.Int)
	submissionTime := values["time"].(*big.Int)
	eventData := RewardsEvent{
		Index:             indexBig,
		ExecutionBlock:    submission.ExecutionBlock,
		ConsensusBlock:    submission.ConsensusBlock,
		IntervalsPassed:   submission.IntervalsPassed,
		TreasuryRPL:       submission.TreasuryRPL,
		TrustedNodeRPL:    submission.TrustedNodeRPL,
		NodeRPL:           submission.NodeRPL,
		NodeETH:           submission.NodeETH,
		UserETH:           submission.UserETH,
		MerkleRoot:        common.BytesToHash(submission.MerkleRoot[:]),
		MerkleTreeCID:     submission.MerkleTreeCID,
		IntervalStartTime: time.Unix(eventIntervalStartTime.Int64(), 0),
		IntervalEndTime:   time.Unix(eventIntervalEndTime.Int64(), 0),
		SubmissionTime:    time.Unix(submissionTime.Int64(), 0),
	}

	return eventData, nil

}

// Get the event info for a rewards snapshot
func GetRewardSnapshotEventWithUpgrades(rp *rocketpool.RocketPool, index uint64, intervalSize *big.Int, startBlock *big.Int, endBlock *big.Int, rocketRewardsPoolAddresses []common.Address, opts *bind.CallOpts) (bool, RewardsEvent, error) {
	// Get contracts
	rocketRewardsPool, err := getRocketRewardsPool(rp, opts)
	if err != nil {
		return false, RewardsEvent{}, err
	}

	rocketRewardsPoolAddresses = append(rocketRewardsPoolAddresses, *rocketRewardsPool.Address)

	// Construct a filter query for relevant logs
	indexBig := big.NewInt(0).SetUint64(index)
	indexBytes := [32]byte{}
	indexBig.FillBytes(indexBytes[:])
	addressFilter := rocketRewardsPoolAddresses
	topicFilter := [][]common.Hash{{rocketRewardsPool.ABI.Events["RewardSnapshot"].ID}, {indexBytes}}

	// Get the event logs
	logs, err := eth.GetLogs(rp, addressFilter, topicFilter, intervalSize, startBlock, endBlock, nil)
	if err != nil {
		return false, RewardsEvent{}, err
	}

	// Get the log info
	values := make(map[string]interface{})
	if len(logs) == 0 {
		return false, RewardsEvent{}, nil
	}
	err = rocketRewardsPool.ABI.Events["RewardSnapshot"].Inputs.UnpackIntoMap(values, logs[0].Data)
	if err != nil {
		return false, RewardsEvent{}, err
	}

	// Get the decoded data
	submissionPrototype := RewardSubmission{}
	submissionType := reflect.TypeOf(submissionPrototype)
	submission := reflect.ValueOf(values["submission"]).Convert(submissionType).Interface().(RewardSubmission)
	eventIntervalStartTime := values["intervalStartTime"].(*big.Int)
	eventIntervalEndTime := values["intervalEndTime"].(*big.Int)
	submissionTime := values["time"].(*big.Int)
	eventData := RewardsEvent{
		Index:             indexBig,
		ExecutionBlock:    submission.ExecutionBlock,
		ConsensusBlock:    submission.ConsensusBlock,
		IntervalsPassed:   submission.IntervalsPassed,
		TreasuryRPL:       submission.TreasuryRPL,
		TrustedNodeRPL:    submission.TrustedNodeRPL,
		NodeRPL:           submission.NodeRPL,
		NodeETH:           submission.NodeETH,
		MerkleRoot:        common.BytesToHash(submission.MerkleRoot[:]),
		MerkleTreeCID:     submission.MerkleTreeCID,
		IntervalStartTime: time.Unix(eventIntervalStartTime.Int64(), 0),
		IntervalEndTime:   time.Unix(eventIntervalEndTime.Int64(), 0),
		SubmissionTime:    time.Unix(submissionTime.Int64(), 0),
	}

	return true, eventData, nil

}

// Get contracts
var rocketRewardsPoolLock sync.Mutex

func getRocketRewardsPool(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketRewardsPoolLock.Lock()
	defer rocketRewardsPoolLock.Unlock()
	return rp.GetContract("rocketRewardsPool", opts)
}
