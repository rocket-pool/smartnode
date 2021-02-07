package rewards

import (
    "fmt"
    "math/big"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
)


// Get whether a claims contract is enabled
func getEnabled(claimsContract *rocketpool.Contract, claimsName string, opts *bind.CallOpts) (bool, error) {
    enabled := new(bool)
    if err := claimsContract.Call(opts, enabled, "getEnabled"); err != nil {
        return false, fmt.Errorf("Could not get %s claims contract enabled status: %w", claimsName, err)
    }
    return *enabled, nil
}


// Get whether a claimer can make a claim
func getClaimPossible(claimsContract *rocketpool.Contract, claimsName string, claimerAddress common.Address, opts *bind.CallOpts) (bool, error) {
    claimPossible := new(bool)
    if err := claimsContract.Call(opts, claimPossible, "getClaimPossible", claimerAddress); err != nil {
        return false, fmt.Errorf("Could not get %s claim possible status for %s: %w", claimsName, claimerAddress.Hex(), err)
    }
    return *claimPossible, nil
}


// Get the percentage of rewards available for a claimer
func getClaimRewardsPerc(claimsContract *rocketpool.Contract, claimsName string, claimerAddress common.Address, opts *bind.CallOpts) (float64, error) {
    claimRewardsPerc := new(*big.Int)
    if err := claimsContract.Call(opts, claimRewardsPerc, "getClaimRewardsPerc", claimerAddress); err != nil {
        return 0, fmt.Errorf("Could not get %s claim rewards percent for %s: %w", claimsName, claimerAddress.Hex(), err)
    }
    return eth.WeiToEth(*claimRewardsPerc), nil
}


// Get the total amount of rewards available for a claimer
func getClaimRewardsAmount(claimsContract *rocketpool.Contract, claimsName string, claimerAddress common.Address, opts *bind.CallOpts) (*big.Int, error) {
    claimRewardsAmount := new(*big.Int)
    if err := claimsContract.Call(opts, claimRewardsAmount, "getClaimRewardsAmount", claimerAddress); err != nil {
        return nil, fmt.Errorf("Could not get %s claim rewards amount for %s: %w", claimsName, claimerAddress.Hex(), err)
    }
    return *claimRewardsAmount, nil
}


// Claim rewards
func claim(claimsContract *rocketpool.Contract, claimsName string, opts *bind.TransactOpts) (*types.Receipt, error) {
    txReceipt, err := claimsContract.Transact(opts, "claim")
    if err != nil {
        return nil, fmt.Errorf("Could not claim %s rewards: %w", claimsName, err)
    }
    return txReceipt, nil
}

