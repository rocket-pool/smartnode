package faucet

import (
    "context"
    "math/big"

    "github.com/urfave/cli"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func canWithdrawRpl(c *cli.Context) (*api.CanFaucetWithdrawRplResponse, error) {

    // Get services
    if err := services.RequireNodeWallet(c); err != nil { return nil, err }
    if err := services.RequireRplFaucet(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    ec, err := services.GetEthClient(c)
    if err != nil { return nil, err }
    f, err := services.GetRplFaucet(c)
    if err != nil { return nil, err }

    // Response
    response := api.CanFaucetWithdrawRplResponse{}

    // Get node account
    nodeAccount, err := w.GetNodeAccount()
    if err != nil {
        return nil, err
    }

    // Data
    var wg errgroup.Group
    var withdrawalFee *big.Int
    var nodeAccountBalance *big.Int

    // Check faucet balance
    wg.Go(func() error {
        balance, err := f.GetBalance(nil)
        if err == nil {
            response.InsufficientFaucetBalance = (balance.Cmp(big.NewInt(0)) == 0)
        }
        return err
    })

    // Check allowance
    wg.Go(func() error {
        allowance, err := f.GetAllowanceFor(nil, nodeAccount.Address)
        if err == nil {
            response.InsufficientAllowance = (allowance.Cmp(big.NewInt(0)) == 0)
        }
        return err
    })

    // Get withdrawal fee
    wg.Go(func() error {
        var err error
        withdrawalFee, err = f.WithdrawalFee(nil)
        return err
    })

    // Get node account balance
    wg.Go(func() error {
        var err error
        nodeAccountBalance, err = ec.BalanceAt(context.Background(), nodeAccount.Address, nil)
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return nil, err
    }

    // Check node account balance
    response.InsufficientNodeBalance = (nodeAccountBalance.Cmp(withdrawalFee) < 0)

    // Update & return response
    response.CanWithdraw = !(response.InsufficientFaucetBalance || response.InsufficientAllowance || response.InsufficientNodeBalance)
    return &response, nil

}


func withdrawRpl(c *cli.Context) (*api.FaucetWithdrawRplResponse, error) {

    // Get services
    if err := services.RequireNodeWallet(c); err != nil { return nil, err }
    if err := services.RequireRplFaucet(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    f, err := services.GetRplFaucet(c)
    if err != nil { return nil, err }

    // Response
    response := api.FaucetWithdrawRplResponse{}

    // Get node account
    nodeAccount, err := w.GetNodeAccount()
    if err != nil {
        return nil, err
    }

    // Data
    var wg errgroup.Group
    var balance *big.Int
    var allowance *big.Int
    var withdrawalFee *big.Int

    // Get faucet balance
    wg.Go(func() error {
        var err error
        balance, err = f.GetBalance(nil)
        return err
    })

    // Get allowance
    wg.Go(func() error {
        var err error
        allowance, err = f.GetAllowanceFor(nil, nodeAccount.Address)
        return err
    })

    // Get withdrawal fee
    wg.Go(func() error {
        var err error
        withdrawalFee, err = f.WithdrawalFee(nil)
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return nil, err
    }

    // Get withdrawal amount
    var amount *big.Int
    if balance.Cmp(allowance) > 0 {
        amount = allowance
    } else {
        amount = balance
    }
    response.Amount = amount

    // Get transactor
    opts, err := w.GetNodeAccountTransactor()
    if err != nil {
        return nil, err
    }
    opts.Value = withdrawalFee

    // Register node
    tx, err := f.Withdraw(opts, amount)
    if err != nil {
        return nil, err
    }
    response.TxHash = tx.Hash()

    // Return response
    return &response, nil

}

