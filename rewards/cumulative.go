package rewards

import (
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
)

// Filters through token claim events and sums the total amount claimed by claimerAddress
func CalculateLifetimeNodeRewards(rp *rocketpool.RocketPool, claimerAddress common.Address, legacyRewardsPoolAddress *common.Address, legacyClaimNodeAddress *common.Address, intervalSize *big.Int, startBlock *big.Int) (*big.Int, error) {
	// Get contracts
	rocketRewardsPool, err := getRocketRewardsPool(rp)
	if err != nil {
		return nil, err
	}
	rocketClaimNode, err := getRocketClaimNode(rp)
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

// Filters through token claim events and sums the total amount claimed by claimerAddress
func CalculateLifetimeTrustedNodeRewards(rp *rocketpool.RocketPool, claimerAddress common.Address, legacyRewardsPoolAddress *common.Address, legacyClaimTrustedNodeAddress *common.Address, intervalSize *big.Int, startBlock *big.Int) (*big.Int, error) {
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

// Get contracts
var rocketClaimNodeLock sync.Mutex
var rocketClaimTrustedNodeLock sync.Mutex

func getRocketClaimNode(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
	rocketClaimNodeLock.Lock()
	defer rocketClaimNodeLock.Unlock()
	return rp.GetContract("rocketClaimNode")
}

func getRocketClaimTrustedNode(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
	rocketClaimTrustedNodeLock.Lock()
	defer rocketClaimTrustedNodeLock.Unlock()
	return rp.GetContract("rocketClaimTrustedNode")
}
