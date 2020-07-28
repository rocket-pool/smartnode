package node

import (
    "context"
    "errors"
    "math/big"

    "github.com/ethereum/go-ethereum/common"
    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/rocket-pool/rocketpool-go/settings"
    "github.com/rocket-pool/rocketpool-go/utils/contract"
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
    am, err := services.GetAccountManager(c)
    if err != nil { return nil, err }
    ec, err := services.GetEthClient(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.CanNodeDepositResponse{}

    // Get node account
    nodeAccount, _ := am.GetNodeAccount()

    // Sync
    var wg errgroup.Group

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
        trusted, err := node.GetNodeTrusted(rp, nodeAccount.Address, nil)
        if err == nil {
            response.InvalidAmount = !trusted
        }
        return err
    })

    // Check node deposits are enabled
    wg.Go(func() error {
        depositEnabled, err := settings.GetNodeRegistrationEnabled(rp, nil)
        if err == nil {
            response.DepositDisabled = !depositEnabled
        }
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return nil, err
    }

    // Update & return response
    response.CanDeposit = !(response.InsufficientBalance || response.InvalidAmount || response.DepositDisabled)
    return &response, nil

}


func nodeDeposit(c *cli.Context, amountWei *big.Int, minNodeFee float64) (*api.NodeDepositResponse, error) {

    // Get services
    if err := services.RequireNodeRegistered(c); err != nil { return nil, err }
    am, err := services.GetAccountManager(c)
    if err != nil { return nil, err }
    ec, err := services.GetEthClient(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.NodeDepositResponse{}

    // Get transactor
    opts, err := am.GetNodeAccountTransactor()
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

    // Get minipool manager contract details
    minipoolManagerAddress, minipoolManagerABI, err := contract.GetDetails(rp, "rocketMinipoolManager")
    if err != nil {
        return nil, err
    }

    // Get created minipool address
    minipoolCreatedEvents, err := contract.GetTransactionEvents(ec, minipoolManagerAddress, minipoolManagerABI, txReceipt, "MinipoolCreated", minipoolCreated{})
    if err != nil || len(minipoolCreatedEvents) == 0 {
        return nil, errors.New("Could not get minipool created event")
    }
    response.MinipoolAddress = minipoolCreatedEvents[0].(minipoolCreated).Minipool

    // Return response
    return &response, nil

}

