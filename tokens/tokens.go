package tokens

import (
    "context"
    "fmt"
    "math/big"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/ethclient"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/utils/contract"
)


// Token balances
type Balances struct {
    ETH *big.Int
    NETH *big.Int
}


// Get token balances of an address
func GetBalances(rp *rocketpool.RocketPool, address common.Address) (*Balances, error) {

    // Data
    var wg errgroup.Group
    var ethBalance *big.Int
    var nethBalance *big.Int

    // Load data
    wg.Go(func() error {
        var err error
        ethBalance, err = rp.Client.BalanceAt(context.Background(), address, nil)
        return err
    })
    wg.Go(func() error {
        var err error
        nethBalance, err = GetNETHBalance(rp, address)
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return nil, err
    }

    // Return
    return &Balances{
        ETH: ethBalance,
        NETH: nethBalance,
    }, nil

}


// Get a token balance
func balanceOf(tokenContract *bind.BoundContract, tokenName string, address common.Address) (*big.Int, error) {
    balance := new(*big.Int)
    if err := tokenContract.Call(nil, balance, "balanceOf", address); err != nil {
        return nil, fmt.Errorf("Could not get %v balance of %v: %w", tokenName, address.Hex(), err)
    }
    return *balance, nil
}


// Transfer tokens to an address
func transfer(client *ethclient.Client, tokenContract *bind.BoundContract, tokenName string, to common.Address, amount *big.Int, opts *bind.TransactOpts) (*types.Receipt, error) {
    txReceipt, err := contract.Transact(client, tokenContract, opts, "transfer", to, amount)
    if err != nil {
        return nil, fmt.Errorf("Could not transfer %v to %v: %w", tokenName, to.Hex(), err)
    }
    return txReceipt, nil
}

