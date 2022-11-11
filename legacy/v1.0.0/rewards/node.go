package rewards

import (
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
)

// Get whether node reward claims are enabled
func GetNodeClaimsEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts, legacyRocketClaimNodeAddress *common.Address) (bool, error) {
	rocketClaimNode, err := getRocketClaimNode(rp, legacyRocketClaimNodeAddress, opts)
	if err != nil {
		return false, err
	}
	return getEnabled(rocketClaimNode, "node", opts)
}

// Get whether a node rewards claimer can claim
func GetNodeClaimPossible(rp *rocketpool.RocketPool, claimerAddress common.Address, opts *bind.CallOpts, legacyRocketClaimNodeAddress *common.Address) (bool, error) {
	rocketClaimNode, err := getRocketClaimNode(rp, legacyRocketClaimNodeAddress, opts)
	if err != nil {
		return false, err
	}
	return getClaimPossible(rocketClaimNode, "node", claimerAddress, opts)
}

// Get the percentage of rewards available for a node rewards claimer
func GetNodeClaimRewardsPerc(rp *rocketpool.RocketPool, claimerAddress common.Address, opts *bind.CallOpts, legacyRocketClaimNodeAddress *common.Address) (float64, error) {
	rocketClaimNode, err := getRocketClaimNode(rp, legacyRocketClaimNodeAddress, opts)
	if err != nil {
		return 0, err
	}
	return getClaimRewardsPerc(rocketClaimNode, "node", claimerAddress, opts)
}

// Get the total amount of rewards available for a node rewards claimer
func GetNodeClaimRewardsAmount(rp *rocketpool.RocketPool, claimerAddress common.Address, opts *bind.CallOpts, legacyRocketClaimNodeAddress *common.Address) (*big.Int, error) {
	rocketClaimNode, err := getRocketClaimNode(rp, legacyRocketClaimNodeAddress, opts)
	if err != nil {
		return nil, err
	}
	return getClaimRewardsAmount(rocketClaimNode, "node", claimerAddress, opts)
}

// Estimate the gas of ClaimNodeRewards
func EstimateClaimNodeRewardsGas(rp *rocketpool.RocketPool, opts *bind.TransactOpts, legacyRocketClaimNodeAddress *common.Address) (rocketpool.GasInfo, error) {
	rocketClaimNode, err := getRocketClaimNode(rp, legacyRocketClaimNodeAddress, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return estimateClaimGas(rocketClaimNode, opts)
}

// Claim node rewards
func ClaimNodeRewards(rp *rocketpool.RocketPool, opts *bind.TransactOpts, legacyRocketClaimNodeAddress *common.Address) (common.Hash, error) {
	rocketClaimNode, err := getRocketClaimNode(rp, legacyRocketClaimNodeAddress, nil)
	if err != nil {
		return common.Hash{}, err
	}
	return claim(rocketClaimNode, "node", opts)
}

// Filters through token claim events and sums the total amount claimed by claimerAddress
func CalculateLifetimeNodeRewards(rp *rocketpool.RocketPool, claimerAddress common.Address, intervalSize *big.Int, startBlock *big.Int, legacyRocketRewardsPoolAddress *common.Address, legacyRocketClaimNodeAddress *common.Address, opts *bind.CallOpts) (*big.Int, error) {
	// Get contracts
	rocketRewardsPool, err := getRocketRewardsPool(rp, legacyRocketRewardsPoolAddress, opts)
	if err != nil {
		return nil, err
	}
	rocketClaimNode, err := getRocketClaimNode(rp, legacyRocketClaimNodeAddress, opts)
	if err != nil {
		return nil, err
	}
	// Construct a filter query for relevant logs
	addressFilter := []common.Address{*rocketRewardsPool.Address}
	// RPLTokensClaimed(address clamingContract, address claimingAddress, uint256 amount, uint256 time)
	topicFilter := [][]common.Hash{{rocketRewardsPool.ABI.Events["RPLTokensClaimed"].ID}, {rocketClaimNode.Address.Hash()}, {claimerAddress.Hash()}}

	// Get the event logs
	logs, err := eth.GetLogs(rp, addressFilter, topicFilter, intervalSize, startBlock, nil, nil)
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

// Get the time that the user registered as a claimer
func GetNodeRegistrationTime(rp *rocketpool.RocketPool, claimerAddress common.Address, opts *bind.CallOpts, legacyRocketRewardsPoolAddress *common.Address) (time.Time, error) {
	return getClaimingContractUserRegisteredTime(rp, "rocketClaimNode", claimerAddress, opts, legacyRocketRewardsPoolAddress)
}

// Get the total rewards claimed for this claiming contract this interval
func GetNodeTotalClaimed(rp *rocketpool.RocketPool, opts *bind.CallOpts, legacyRocketRewardsPoolAddress *common.Address) (*big.Int, error) {
	return getClaimingContractTotalClaimed(rp, "rocketClaimNode", opts, legacyRocketRewardsPoolAddress)
}

// Get contracts
var rocketClaimNodeLock sync.Mutex

func getRocketClaimNode(rp *rocketpool.RocketPool, address *common.Address, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketClaimNodeLock.Lock()
	defer rocketClaimNodeLock.Unlock()
	if address == nil {
		return rp.VersionManager.V1_0_0.GetContract("rocketClaimNode", opts)
	} else {
		return rp.VersionManager.V1_0_0.GetContractWithAddress("rocketClaimNode", *address)
	}
}
