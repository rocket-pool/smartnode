package settings

import (
    "fmt"
    "math/big"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/core/types"

    "github.com/rocket-pool/rocketpool-go/dao/protocol"
    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
)


// Config
const RewardsSettingsContractName = "rocketDAOProtocolSettingsRewards"


// The claim amount for a claimer as a fraction
func GetRewardsClaimerPerc(rp *rocketpool.RocketPool, contractName string, opts *bind.CallOpts) (float64, error) {
    rewardsSettingsContract, err := getRewardsSettingsContract(rp)
    if err != nil {
        return 0, err
    }
    value := new(*big.Int)
    if err := rewardsSettingsContract.Call(opts, value, "getRewardsClaimerPerc", contractName); err != nil {
        return 0, fmt.Errorf("Could not get rewards claimer percent: %w", err)
    }
    return eth.WeiToEth(*value), nil
}


// The block that a claimer's share was last updated at
func GetRewardsClaimerPercBlockUpdated(rp *rocketpool.RocketPool, contractName string, opts *bind.CallOpts) (uint64, error) {
    rewardsSettingsContract, err := getRewardsSettingsContract(rp)
    if err != nil {
        return 0, err
    }
    value := new(*big.Int)
    if err := rewardsSettingsContract.Call(opts, value, "getRewardsClaimerPercBlockUpdated", contractName); err != nil {
        return 0, fmt.Errorf("Could not get rewards claimer updated block: %w", err)
    }
    return (*value).Uint64(), nil
}


// The total claim amount for all claimers as a fraction
func GetRewardsClaimersPercTotal(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
    rewardsSettingsContract, err := getRewardsSettingsContract(rp)
    if err != nil {
        return 0, err
    }
    value := new(*big.Int)
    if err := rewardsSettingsContract.Call(opts, value, "getRewardsClaimersPercTotal"); err != nil {
        return 0, fmt.Errorf("Could not get rewards claimers total percent: %w", err)
    }
    return eth.WeiToEth(*value), nil
}


// Rewards claim interval in blocks
func GetRewardsClaimIntervalBlocks(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
    rewardsSettingsContract, err := getRewardsSettingsContract(rp)
    if err != nil {
        return 0, err
    }
    value := new(*big.Int)
    if err := rewardsSettingsContract.Call(opts, value, "getRewardsClaimIntervalBlocks"); err != nil {
        return 0, fmt.Errorf("Could not get rewards claim interval: %w", err)
    }
    return (*value).Uint64(), nil
}
func BootstrapRewardsClaimIntervalBlocks(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (*types.Receipt, error) {
    return protocol.BootstrapUint(rp, RewardsSettingsContractName, "rpl.rewards.claim.period.blocks", big.NewInt(int64(value)), opts)
}


// Get contracts
var rewardsSettingsContractLock sync.Mutex
func getRewardsSettingsContract(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
    rewardsSettingsContractLock.Lock()
    defer rewardsSettingsContractLock.Unlock()
    return rp.GetContract(RewardsSettingsContractName)
}

