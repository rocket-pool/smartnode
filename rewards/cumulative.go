package rewards

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
)

// Filters through token claim events and sums the total amount claimed by claimerAddress
func CalculateLegacyLifetimeNodeRewards(rp *rocketpool.RocketPool, claimerAddress common.Address, legacyRewardsPoolAddress common.Address, legacyRewardsPoolAbi string, legacyClaimNodeAddress common.Address, intervalSize *big.Int, startBlock *big.Int) (*big.Int, error) {
	// Create RocketRewardsPool ABI
	abi, err := rocketpool.DecodeAbi(legacyRewardsPoolAbi)
	if err != nil {
		return nil, fmt.Errorf("error decoding legacy RocketRewardsPool ABI: %w", err)
	}

	// Create RocketRewardsPool contract
	rocketRewardsPool := &rocketpool.Contract{
		Contract: bind.NewBoundContract(legacyRewardsPoolAddress, *abi, rp.Client, rp.Client, rp.Client),
		Address:  &legacyRewardsPoolAddress,
		ABI:      abi,
		Client:   rp.Client,
	}

	// Construct a filter query for relevant logs
	addressFilter := []common.Address{legacyRewardsPoolAddress}
	// RPLTokensClaimed(address clamingContract, address claimingAddress, uint256 amount, uint256 time)
	topicFilter := [][]common.Hash{{rocketRewardsPool.ABI.Events["RPLTokensClaimed"].ID}, {legacyClaimNodeAddress.Hash()}, {claimerAddress.Hash()}}

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
func CalculateLegacyLifetimeTrustedNodeRewards(rp *rocketpool.RocketPool, claimerAddress common.Address, legacyRewardsPoolAddress common.Address, legacyRewardsPoolAbi string, legacyClaimTrustedNodeAddress common.Address, intervalSize *big.Int, startBlock *big.Int) (*big.Int, error) {
	// Create RocketRewardsPool ABI
	abi, err := rocketpool.DecodeAbi(legacyRewardsPoolAbi)
	if err != nil {
		return nil, fmt.Errorf("error decoding legacy RocketRewardsPool ABI: %w", err)
	}

	// Create RocketRewardsPool contract
	rocketRewardsPool := &rocketpool.Contract{
		Contract: bind.NewBoundContract(legacyRewardsPoolAddress, *abi, rp.Client, rp.Client, rp.Client),
		Address:  &legacyRewardsPoolAddress,
		ABI:      abi,
		Client:   rp.Client,
	}

	// Construct a filter query for relevant logs
	addressFilter := []common.Address{*rocketRewardsPool.Address}
	// RPLTokensClaimed(address clamingContract, address clainingAddress, uint256 amount, uint256 time)
	topicFilter := [][]common.Hash{{rocketRewardsPool.ABI.Events["RPLTokensClaimed"].ID}, {legacyClaimTrustedNodeAddress.Hash()}, {claimerAddress.Hash()}}

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
