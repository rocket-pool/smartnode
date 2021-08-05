package rewards

import (
    "context"
    "github.com/ethereum/go-ethereum"
    "math/big"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
)

// Get whether trusted node reward claims are enabled
func GetTrustedNodeClaimsEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
    rocketClaimTrustedNode, err := getRocketClaimTrustedNode(rp)
    if err != nil {
        return false, err
    }
    return getEnabled(rocketClaimTrustedNode, "trusted node", opts)
}


// Get whether a trusted node rewards claimer can claim
func GetTrustedNodeClaimPossible(rp *rocketpool.RocketPool, claimerAddress common.Address, opts *bind.CallOpts) (bool, error) {
    rocketClaimTrustedNode, err := getRocketClaimTrustedNode(rp)
    if err != nil {
        return false, err
    }
    return getClaimPossible(rocketClaimTrustedNode, "trusted node", claimerAddress, opts)
}


// Get the percentage of rewards available for a trusted node rewards claimer
func GetTrustedNodeClaimRewardsPerc(rp *rocketpool.RocketPool, claimerAddress common.Address, opts *bind.CallOpts) (float64, error) {
    rocketClaimTrustedNode, err := getRocketClaimTrustedNode(rp)
    if err != nil {
        return 0, err
    }
    return getClaimRewardsPerc(rocketClaimTrustedNode, "trusted node", claimerAddress, opts)
}


// Get the total amount of rewards available for a trusted node rewards claimer
func GetTrustedNodeClaimRewardsAmount(rp *rocketpool.RocketPool, claimerAddress common.Address, opts *bind.CallOpts) (*big.Int, error) {
    rocketClaimTrustedNode, err := getRocketClaimTrustedNode(rp)
    if err != nil {
        return nil, err
    }
    return getClaimRewardsAmount(rocketClaimTrustedNode, "trusted node", claimerAddress, opts)
}


// Estimate the gas of ClaimTrustedNodeRewards
func EstimateClaimTrustedNodeRewardsGas(rp *rocketpool.RocketPool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
    rocketClaimTrustedNode, err := getRocketClaimTrustedNode(rp)
    if err != nil {
        return rocketpool.GasInfo{}, err
    }
    return estimateClaimGas(rocketClaimTrustedNode, opts)
}


// Claim trusted node rewards
func ClaimTrustedNodeRewards(rp *rocketpool.RocketPool, opts *bind.TransactOpts) (common.Hash, error) {
    rocketClaimTrustedNode, err := getRocketClaimTrustedNode(rp)
    if err != nil {
        return common.Hash{}, err
    }
    return claim(rocketClaimTrustedNode, "trusted node", opts)
}


// Filters through token claim events and sums the total amount claimed by claimerAddress
func CalculateLifetimeTrustedNodeRewards(rp *rocketpool.RocketPool, claimerAddress common.Address) (*big.Int, error) {
    // Get contracts
    rocketRewardsPool, err := getRocketRewardsPool(rp)
    if err != nil {
        return nil, err
    }
    rocketClaimTrustedNode, err := getRocketClaimTrustedNode(rp)
    if err != nil {
        return nil, err
    }
    // Construct a filter query for relevant logs
    addressFilter := []common.Address{*rocketRewardsPool.Address}
    // RPLTokensClaimed(address clamingContract, address clainingAddress, uint256 amount, uint256 time)
    topicFilter := [][]common.Hash{{rocketRewardsPool.ABI.Events["RPLTokensClaimed"].ID}, {rocketClaimTrustedNode.Address.Hash()}, {claimerAddress.Hash()}}
    logs, err := rp.Client.FilterLogs(context.Background(), ethereum.FilterQuery{
        Addresses: addressFilter,
        Topics: topicFilter,
    })
    if err != nil {
        return nil, err
    }
    // Iterate over the logs and sum the amount
    sum := big.NewInt(0)
    for _, log := range logs {
        values := make(map[string]interface{})
        // Decode the event
        if rocketRewardsPool.ABI.Events["RPLTokensClaimed"].Inputs.UnpackIntoMap(values, log.Data) != nil {
            return nil, err
        }
        // Add the amount argument to our sum
        amount := values["amount"].(*big.Int)
        sum.Add(sum, amount)
    }
    // Return the result
    return sum, nil
}

// Get contracts
var rocketClaimTrustedNodeLock sync.Mutex
func getRocketClaimTrustedNode(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
    rocketClaimTrustedNodeLock.Lock()
    defer rocketClaimTrustedNodeLock.Unlock()
    return rp.GetContract("rocketClaimTrustedNode")
}

