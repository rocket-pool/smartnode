package rewards

import (
    "math/big"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"

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


// Claim trusted node rewards
func ClaimTrustedNodeRewards(rp *rocketpool.RocketPool, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketClaimTrustedNode, err := getRocketClaimTrustedNode(rp)
    if err != nil {
        return nil, err
    }
    return claim(rocketClaimTrustedNode, "trusted node", opts)
}


// Get contracts
var rocketClaimTrustedNodeLock sync.Mutex
func getRocketClaimTrustedNode(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
    rocketClaimTrustedNodeLock.Lock()
    defer rocketClaimTrustedNodeLock.Unlock()
    return rp.GetContract("rocketClaimTrustedNode")
}

