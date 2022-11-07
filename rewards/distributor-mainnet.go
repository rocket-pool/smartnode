package rewards

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

// Check if the given node has already claimed rewards for the given interval
func IsClaimed(rp *rocketpool.RocketPool, index *big.Int, claimerAddress common.Address, opts *bind.CallOpts) (bool, error) {
	rocketDistributorMainnet, err := getRocketDistributorMainnet(rp, opts)
	if err != nil {
		return false, err
	}
	isClaimed := new(bool)
	if err := rocketDistributorMainnet.Call(opts, isClaimed, "isClaimed", index, claimerAddress); err != nil {
		return false, fmt.Errorf("Could not get rewards claim status for interval %s, node %s: %w", index.String(), claimerAddress.Hex(), err)
	}
	return *isClaimed, nil
}

// Get the Merkle root for an interval
func MerkleRoots(rp *rocketpool.RocketPool, interval *big.Int, opts *bind.CallOpts) ([]byte, error) {
	rocketDistributorMainnet, err := getRocketDistributorMainnet(rp, opts)
	if err != nil {
		return nil, err
	}
	bytes := new([32]byte)
	if err := rocketDistributorMainnet.Call(opts, bytes, "merkleRoots", interval); err != nil {
		return nil, fmt.Errorf("Could not get Merkle root for interval %s: %w", interval.String(), err)
	}
	return (*bytes)[:], nil
}

// Estimate claim rewards gas
func EstimateClaimGas(rp *rocketpool.RocketPool, address common.Address, indices []*big.Int, amountRPL []*big.Int, amountETH []*big.Int, merkleProofs [][]common.Hash, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDistributorMainnet, err := getRocketDistributorMainnet(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDistributorMainnet.GetTransactionGasInfo(opts, "claim", address, indices, amountRPL, amountETH, merkleProofs)
}

// Claim rewards
func Claim(rp *rocketpool.RocketPool, address common.Address, indices []*big.Int, amountRPL []*big.Int, amountETH []*big.Int, merkleProofs [][]common.Hash, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDistributorMainnet, err := getRocketDistributorMainnet(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketDistributorMainnet.Transact(opts, "claim", address, indices, amountRPL, amountETH, merkleProofs)
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not claim rewards: %w", err)
	}
	return tx.Hash(), nil
}

// Estimate claim and restake rewards gas
func EstimateClaimAndStakeGas(rp *rocketpool.RocketPool, address common.Address, indices []*big.Int, amountRPL []*big.Int, amountETH []*big.Int, merkleProofs [][]common.Hash, stakeAmount *big.Int, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDistributorMainnet, err := getRocketDistributorMainnet(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDistributorMainnet.GetTransactionGasInfo(opts, "claimAndStake", address, indices, amountRPL, amountETH, merkleProofs, stakeAmount)
}

// Claim and restake rewards
func ClaimAndStake(rp *rocketpool.RocketPool, address common.Address, indices []*big.Int, amountRPL []*big.Int, amountETH []*big.Int, merkleProofs [][]common.Hash, stakeAmount *big.Int, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDistributorMainnet, err := getRocketDistributorMainnet(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketDistributorMainnet.Transact(opts, "claimAndStake", address, indices, amountRPL, amountETH, merkleProofs, stakeAmount)
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not claim rewards: %w", err)
	}
	return tx.Hash(), nil
}

// Get contracts
var rocketDistributorMainnetLock sync.Mutex

func getRocketDistributorMainnet(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketDistributorMainnetLock.Lock()
	defer rocketDistributorMainnetLock.Unlock()
	return rp.GetContract("rocketMerkleDistributorMainnet", opts)
}
