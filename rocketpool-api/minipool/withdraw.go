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
    if !canWithdraw.Success {
        var message string
        if canWithdraw.MinipoolDidNotExist {
            message = "The specified minipool does not exist"
        } else if canWithdraw.WithdrawalsDisabled {
            message = "Minipool withdrawals are currently disabled in Rocket Pool"
        } else if canWithdraw.InvalidNodeOwner {
            message = "The specified minipool is not owned by the node"
        } else if canWithdraw.InvalidStatus {
            message = "The specified minipool is not ready for withdrawal"
        } else if canWithdraw.NodeDepositDidNotExist {
            message = "The node deposit has already been withdrawn from the specified minipool"
        }
        api.PrintResponse(p.Output, canWithdraw, message)
        return nil
    }

    // Withdraw node deposit from minipool
    withdrawn, err := minipool.WithdrawMinipool(p, minipoolAddress)
    if err != nil { return err }

    // Print response
    api.PrintResponse(p.Output, withdrawn, "")
    return nil

}

