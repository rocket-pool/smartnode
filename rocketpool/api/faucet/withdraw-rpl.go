package faucet

import (
	"context"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/client"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/contracts"
	"github.com/rocket-pool/smartnode/shared/types/api"
)


func canWithdrawRpl(c *cli.Context) (*api.CanFaucetWithdrawRplResponse, error) {

    // Get services
    if err := services.RequireNodeWallet(c); err != nil { return nil, err }
    if err := services.RequireRplFaucet(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    ec, err := services.GetEthClientProxy(c)
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
    var balance *big.Int
    var allowance *big.Int

    // Check faucet balance
    wg.Go(func() error {
        _balance, err := f.GetBalance(nil)
        if err == nil {
            response.InsufficientFaucetBalance = (_balance.Cmp(big.NewInt(0)) == 0)
        }
        balance = _balance
        return err
    })

    // Check allowance
    wg.Go(func() error {
        _allowance, err := f.GetAllowanceFor(nil, nodeAccount.Address)
        if err == nil {
            response.InsufficientAllowance = (_allowance.Cmp(big.NewInt(0)) == 0)
        }
        allowance = _allowance
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
    
    if response.CanWithdraw {
        // Get the gas estimate
        opts, err := w.GetNodeAccountTransactor()
        if err != nil {
            return nil, err
        }
        opts.Value = withdrawalFee
        
        // Get withdrawal amount
        var amount *big.Int
        if balance.Cmp(allowance) > 0 {
            amount = allowance
        } else {
            amount = balance
        }

        gasInfo, err := estimateWithdrawGas(c, ec, f, opts, amount)
        if err != nil {
            return nil, err
        }
        response.GasInfo = gasInfo
    }
    
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


func estimateWithdrawGas(c *cli.Context, client *client.EthClientProxy, faucet *contracts.RPLFaucet, opts *bind.TransactOpts, amount *big.Int) (rocketpool.GasInfo, error) {

    response := rocketpool.GasInfo{}

    // Get the faucet address
    config, err := services.GetConfig(c)
    if err != nil {
        return response, err
    }
    faucetAddress := common.HexToAddress(config.Rocketpool.RPLFaucetAddress)

    // Create a contract for the faucet
    faucetAbi, err := abi.JSON(strings.NewReader(contracts.RPLFaucetABI))
    if err != nil {
        return response, err
    }
    contract := &rocketpool.Contract{
        Contract: bind.NewBoundContract(faucetAddress, faucetAbi, client, client, client),
        Address: &faucetAddress,
        ABI: &faucetAbi,
        Client: client,
    }

    // Get the gas info
    gasInfo, err := contract.GetTransactionGasInfo(opts, "withdraw", amount)
    if err != nil {
        return response, err
    }
    return gasInfo, nil

}

