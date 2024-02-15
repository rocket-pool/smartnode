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
	RewardsSettingsContractName            string = "rocketDAOProtocolSettingsRewards"
	RewardsClaimIntervalPeriodsSettingPath string = "dao.protocol.rewards.claimsperiods"
)

// Rewards claimer percents
type RplRewardsPercentages struct {
	OdaoPercentage *big.Int `abi:"trustedNodePerc"`
	PdaoPercentage *big.Int `abi:"protocolPerc"`
	NodePercentage *big.Int `abi:"nodePerc"`
}

// The RPL rewards percentages for the Oracle DAO, Protocol DAO, and node operators
func GetRewardsPercentages(rp *rocketpool.RocketPool, opts *bind.CallOpts) (RplRewardsPercentages, error) {
	rewardsSettingsContract, err := getRewardsSettingsContract(rp, opts)
	if err != nil {
		return RplRewardsPercentages{}, err
	}
	value := new(RplRewardsPercentages)
	if err := rewardsSettingsContract.Call(opts, value, "getRewardsClaimersPerc"); err != nil {
		return RplRewardsPercentages{}, fmt.Errorf("error getting rewards percentages: %w", err)
	}
	return *value, nil
}

// The total RPL rewards percentage for node operator collateral
func GetNodeOperatorRewardsPercent(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rewardsSettingsContract, err := getRewardsSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := rewardsSettingsContract.Call(opts, value, "getRewardsClaimersNodePerc"); err != nil {
		return nil, fmt.Errorf("error getting node operator rewards percent: %w", err)
	}
	return *value, nil
}

// The total RPL rewards percentage for Oracle DAO members
func GetOracleDAORewardsPercent(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rewardsSettingsContract, err := getRewardsSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := rewardsSettingsContract.Call(opts, value, "getRewardsClaimersTrustedNodePerc"); err != nil {
		return nil, fmt.Errorf("error getting oracle DAO rewards percent: %w", err)
	}
	return *value, nil
}

// The total RPL rewards percentage for the Protocol DAO treasury
func GetProtocolDAORewardsPercent(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rewardsSettingsContract, err := getRewardsSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := rewardsSettingsContract.Call(opts, value, "getRewardsClaimersProtocolPerc"); err != nil {
		return nil, fmt.Errorf("error getting protocol DAO rewards percent: %w", err)
	}
	return *value, nil
}

// The time that the RPL rewards percentages were last updated
func GetRewardsClaimerPercTimeUpdated(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	rewardsSettingsContract, err := getRewardsSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := rewardsSettingsContract.Call(opts, value, "getRewardsClaimersTimeUpdated"); err != nil {
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
func ProposeRewardsClaimIntervalTime(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", RewardsClaimIntervalPeriodsSettingPath), RewardsSettingsContractName, RewardsClaimIntervalPeriodsSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeRewardsClaimIntervalTimeGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", RewardsClaimIntervalPeriodsSettingPath), RewardsSettingsContractName, RewardsClaimIntervalPeriodsSettingPath, value, blockNumber, treeNodes, opts)
}

// Get contracts
var rewardsSettingsContractLock sync.Mutex

func getRewardsSettingsContract(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rewardsSettingsContractLock.Lock()
	defer rewardsSettingsContractLock.Unlock()
	return rp.GetContract(RewardsSettingsContractName, opts)
}
