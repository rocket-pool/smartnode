package auction

import (
    "math/big"

    "github.com/rocket-pool/rocketpool-go/auction"
    "github.com/rocket-pool/rocketpool-go/network"
    "github.com/rocket-pool/rocketpool-go/settings/protocol"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
    "github.com/urfave/cli"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func canCreateLot(c *cli.Context) (*api.CanCreateLotResponse, error) {

    // Get services
    if err := services.RequireNodeWallet(c); err != nil { return nil, err }
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.CanCreateLotResponse{}

    // Data
    var wg errgroup.Group
    var remainingRplBalance *big.Int
    var lotMinimumEthValue *big.Int
    var rplPrice *big.Int

    // Check if lot creation is enabled
    wg.Go(func() error {
        createLotEnabled, err := protocol.GetCreateLotEnabled(rp, nil)
        if err == nil {
            response.CreateLotDisabled = !createLotEnabled
        }
        return err
    })

    // Get data
    wg.Go(func() error {
        var err error
        remainingRplBalance, err = auction.GetRemainingRPLBalance(rp, nil)
        return err
    })
    wg.Go(func() error {
        var err error
        lotMinimumEthValue, err = protocol.GetLotMinimumEthValue(rp, nil)
        return err
    })
    wg.Go(func() error {
        var err error
        rplPrice, err = network.GetRPLPrice(rp, nil)
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return nil, err
    }

    // Check auction contract remaining RPL balance
    var tmp big.Int
    var lotMinimumRplAmount big.Int
    tmp.Mul(lotMinimumEthValue, eth.EthToWei(1))
    lotMinimumRplAmount.Quo(&tmp, rplPrice)
    response.InsufficientBalance = (remainingRplBalance.Cmp(&lotMinimumRplAmount) < 0)

    // Update & return response
    response.CanCreate = !(response.InsufficientBalance || response.CreateLotDisabled)
    return &response, nil

}


func createLot(c *cli.Context) (*api.CreateLotResponse, error) {

    // Get services
    if err := services.RequireNodeWallet(c); err != nil { return nil, err }
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.CreateLotResponse{}

    // Get transactor
    opts, err := w.GetNodeAccountTransactor()
    if err != nil {
        return nil, err
    }

    // Create lot
    lotIndex, txReceipt, err := auction.CreateLot(rp, opts)
    if err != nil {
        return nil, err
    }
    response.LotId = lotIndex
    response.TxHash = txReceipt.TxHash

    // Return response
    return &response, nil

}

