package rewards

import (
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

// Get whether node reward claims are enabled
func GetNodeClaimsEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
    rocketClaimNode, err := getRocketClaimNode(rp)
    if err != nil {
        return false, err
    }
    return getEnabled(rocketClaimNode, "node", opts)
}


// Get whether a node rewards claimer can claim
func GetNodeClaimPossible(rp *rocketpool.RocketPool, claimerAddress common.Address, opts *bind.CallOpts) (bool, error) {
    rocketClaimNode, err := getRocketClaimNode(rp)
    if err != nil {
        return false, err
    }
    return getClaimPossible(rocketClaimNode, "node", claimerAddress, opts)
}


// Get the percentage of rewards available for a node rewards claimer
func GetNodeClaimRewardsPerc(rp *rocketpool.RocketPool, claimerAddress common.Address, opts *bind.CallOpts) (float64, error) {
    rocketClaimNode, err := getRocketClaimNode(rp)
    if err != nil {
        return 0, err
    }
    return getClaimRewardsPerc(rocketClaimNode, "node", claimerAddress, opts)
}


// Get the total amount of rewards available for a node rewards claimer
func GetNodeClaimRewardsAmount(rp *rocketpool.RocketPool, claimerAddress common.Address, opts *bind.CallOpts) (*big.Int, error) {
    rocketClaimNode, err := getRocketClaimNode(rp)
    if err != nil {
        return nil, err
    }
    return getClaimRewardsAmount(rocketClaimNode, "node", claimerAddress, opts)
}


// Claim node rewards
func ClaimNodeRewards(rp *rocketpool.RocketPool, opts *bind.TransactOpts) (common.Hash, error) {
    rocketClaimNode, err := getRocketClaimNode(rp)
    if err != nil {
        return common.Hash{}, err
    }
    return claim(rocketClaimNode, "node", opts)
}


// Get contracts
var rocketClaimNodeLock sync.Mutex
func getRocketClaimNode(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
    rocketClaimNodeLock.Lock()
    defer rocketClaimNodeLock.Unlock()
    return rp.GetContract("rocketClaimNode")
}

