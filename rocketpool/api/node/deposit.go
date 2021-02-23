package node

import (
    "context"
    "errors"
    "math/big"

    "github.com/ethereum/go-ethereum/common"
    "github.com/rocket-pool/rocketpool-go/dao/trustednode"
    "github.com/rocket-pool/rocketpool-go/minipool"
    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/rocket-pool/rocketpool-go/settings/protocol"
    "github.com/urfave/cli"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


type minipoolCreated struct {
    Minipool common.Address
    Node common.Address
    Time *big.Int
}


func canNodeDeposit(c *cli.Context, amountWei *big.Int) (*api.CanNodeDepositResponse, error) {

    // Get services
    if err := services.RequireNodeRegistered(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    ec, err := services.GetEthClient(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.CanNodeDepositResponse{}

    // Get node account
    nodeAccount, err := w.GetNodeAccount()
    if err != nil {
        return nil, err
    }

    // Data
    var wg errgroup.Group
    var minipoolCount uint64
    var minipoolLimit uint64

    // Check node balance
    wg.Go(func() error {
        ethBalanceWei, err := ec.BalanceAt(context.Background(), nodeAccount.Address, nil)
        if err == nil {
            response.InsufficientBalance = (amountWei.Cmp(ethBalanceWei) > 0)
        }
        return err
    })

    // Check deposit amount
    wg.Go(func() error {
        if amountWei.Cmp(big.NewInt(0)) > 0 {
            return nil
        }
        trusted, err := trustednode.GetMemberExists(rp, nodeAccount.Address, nil)
        if err == nil {
            response.InvalidAmount = !trusted
        }
        return err
    })

    // Check node deposits are enabled
    wg.Go(func() error {
        depositEnabled, err := protocol.GetNodeDepositEnabled(rp, nil)
        if err == nil {
            response.DepositDisabled = !depositEnabled
        }
        return err
    })

    // Get node staking information
    wg.Go(func() error {
        var err error
        minipoolCount, err = minipool.GetNodeMinipoolCount(rp, nodeAccount.Address, nil)
        return err
    })
    wg.Go(func() error {
        var err error
        minipoolLimit, err = node.GetNodeMinipoolLimit(rp, nodeAccount.Address, nil)
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return nil, err
    }

    // Check node RPL stake
    response.InsufficientRplStake = (minipoolCount >= minipoolLimit)

    // Update & return response
    response.CanDeposit = !(response.InsufficientBalance || response.InsufficientRplStake || response.InvalidAmount || response.DepositDisabled)
    return &response, nil

}


func nodeDeposit(c *cli.Context, amountWei *big.Int, minNodeFee float64) (*api.NodeDepositResponse, error) {

    // Get services
    if err := services.RequireNodeRegistered(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.NodeDepositResponse{}

    // Get transactor
    opts, err := w.GetNodeAccountTransactor()
    if err != nil {
        return nil, err
    }
    opts.Value = amountWei

    // Deposit
    txReceipt, err := node.Deposit(rp, minNodeFee, opts)
    if err != nil {
        return nil, err
    }
    response.TxHash = txReceipt.TxHash

    // Get minipool manager contract
    minipoolManager, err := rp.GetContract("rocketMinipoolManager")
    if err != nil {
        return nil, err
    }

    // Get created minipool address
    minipoolCreatedEvents, err := minipoolManager.GetTransactionEvents(txReceipt, "MinipoolCreated", minipoolCreated{})
    if err != nil || len(minipoolCreatedEvents) == 0 {
        return nil, errors.New("Could not get minipool created event")
    }
    response.MinipoolAddress = minipoolCreatedEvents[0].(minipoolCreated).Minipool

    // Return response
    return &response, nil

}

