package node

import (
    "github.com/ethereum/go-ethereum/common"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/api/node"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/api"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Send resources from the node account to an address
func sendFromNode(c *cli.Context, address string, amount float64, unit string) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        Client: true,
        CM: true,
        LoadContracts: []string{"rocketETHToken", "rocketPoolToken"},
        ClientConn: true,
        ClientSync: true,
        RocketStorage: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Get args
    toAddress := common.HexToAddress(address)
    amountWei := eth.EthToWei(amount)

    // Check tokens can be sent from node
    canSend, err := node.CanSendFromNode(p, amountWei, unit)
    if err != nil { return err }

    // Check response
    if !canSend.Success {
        var message string
        if canSend.InsufficientAccountBalance {
            message = "Node account has insufficient balance for transfer"
        }
        api.PrintResponse(p.Output, canSend, message)
        return nil
    }

    // Send from node
    sent, err := node.SendFromNode(p, toAddress, amountWei, unit)
    if err != nil { return err }

    // Print response
    api.PrintResponse(p.Output, sent, "")
    return nil

}

