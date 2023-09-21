package protocol

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
)

// Config
const (
	RewardsSettingsContractName         string = "rocketDAOProtocolSettingsRewards"
	RewardsClaimIntervalTimeSettingPath string = "rpl.rewards.claim.period.time"
)

// The claim amount for a claimer as a fraction
func GetRewardsClaimerPerc(rp *rocketpool.RocketPool, contractName string, opts *bind.CallOpts) (float64, error) {
	rewardsSettingsContract, err := getRewardsSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := rewardsSettingsContract.Call(opts, value, "getRewardsClaimerPerc", contractName); err != nil {
		return 0, fmt.Errorf("error getting rewards claimer percent: %w", err)
	}
	return eth.WeiToEth(*value), nil
}

// The time that a claimer's share was last updated
func GetRewardsClaimerPercTimeUpdated(rp *rocketpool.RocketPool, contractName string, opts *bind.CallOpts) (uint64, error) {
	rewardsSettingsContract, err := getRewardsSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := rewardsSettingsContract.Call(opts, value, "getRewardsClaimerPercTimeUpdated", contractName); err != nil {
		return 0, fmt.Errorf("error getting rewards claimer updated time: %w", err)
	}
	return (*value).Uint64(), nil
}

// The total claim amount for all claimers as a fraction
func GetRewardsClaimersPercTotal(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
	rewardsSettingsContract, err := getRewardsSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := rewardsSettingsContract.Call(opts, value, "getRewardsClaimersPercTotal"); err != nil {
		return 0, fmt.Errorf("error getting rewards claimers total percent: %w", err)
	}
	return eth.WeiToEth(*value), nil
}

// Rewards claim interval time
func GetRewardsClaimIntervalTime(rp *rocketpool.RocketPool, opts *bind.CallOpts) (time.Duration, error) {
	rewardsSettingsContract, err := getRewardsSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := rewardsSettingsContract.Call(opts, value, "getRewardsClaimIntervalTime"); err != nil {
		return 0, fmt.Errorf("error getting rewards claim interval: %w", err)
	}
	return time.Duration((*value).Uint64()) * time.Second, nil
}
func ProposeRewardsClaimIntervalTime(rp *rocketpool.RocketPool, value time.Duration, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", RewardsClaimIntervalTimeSettingPath), RewardsSettingsContractName, RewardsClaimIntervalTimeSettingPath, big.NewInt(0).SetUint64(uint64(value.Seconds())), blockNumber, treeNodes, opts)
}
func EstimateProposeRewardsClaimIntervalTimeGas(rp *rocketpool.RocketPool, value time.Duration, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", RewardsClaimIntervalTimeSettingPath), RewardsSettingsContractName, RewardsClaimIntervalTimeSettingPath, big.NewInt(0).SetUint64(uint64(value.Seconds())), blockNumber, treeNodes, opts)
}

// Get contracts
var rewardsSettingsContractLock sync.Mutex

func getRewardsSettingsContract(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rewardsSettingsContractLock.Lock()
	defer rewardsSettingsContractLock.Unlock()
	return rp.GetContract(RewardsSettingsContractName, opts)
}
