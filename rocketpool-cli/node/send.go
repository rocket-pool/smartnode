package node

import (
    "fmt"

    "github.com/ethereum/go-ethereum/common"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/api/node"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Send resources from the node to an address
func sendFromNode(c *cli.Context, address string, amount float64, unit string) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        Client: true,
        CM: true,
        LoadContracts: []string{"rocketETHToken", "rocketPoolToken"},
        WaitClientSync: true,
        WaitRocketStorage: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Send from node
    response, err := node.SendFromNode(p, common.HexToAddress(address), eth.EthToWei(amount), unit)
    if err != nil { return err }

    // Print output & return
    if response.InsufficientAccountBalance {
        fmt.Fprintln(p.Output, fmt.Sprintf("Send amount exceeds node account %s balance", unit))
    }
    if response.Success {
        fmt.Fprintln(p.Output, fmt.Sprintf("Successfully sent %.2f %s from node account to %s", amount, unit, address))
    }
    return nil

}

