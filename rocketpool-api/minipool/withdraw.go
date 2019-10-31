package minipool

import (
    "github.com/ethereum/go-ethereum/common"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/api/minipool"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/api"
)


// Withdraw node deposit from minipool
func withdrawMinipool(c *cli.Context, address string) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        Client: true,
        CM: true,
        NodeContractAddress: true,
        LoadContracts: []string{"rocketNodeAPI", "rocketNodeSettings"},
        LoadAbis: []string{"rocketMinipool", "rocketMinipoolDelegateNode", "rocketNodeContract"},
        ClientConn: true,
        ClientSync: true,
        RocketStorage: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Get args
    minipoolAddress := common.HexToAddress(address)

    // Check node deposit can be withdrawn from minipool
    canWithdraw, err := minipool.CanWithdrawMinipool(p, minipoolAddress)
    if err != nil { return err }

    // Check response
    if canWithdraw.MinipoolDidNotExist || canWithdraw.WithdrawalsDisabled || canWithdraw.InvalidNodeOwner || canWithdraw.InvalidStatus || canWithdraw.NodeDepositDidNotExist {
        api.PrintResponse(p.Output, canWithdraw)
        return nil
    }

    // Withdraw node deposit from minipool
    withdrawn, err := minipool.WithdrawMinipool(p, minipoolAddress)
    if err != nil { return err }

    // Print response
    api.PrintResponse(p.Output, withdrawn)
    return nil

}

