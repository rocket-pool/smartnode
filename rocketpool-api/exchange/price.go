package exchange

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/api/exchange"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/api"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Get the current token price
func getTokenPrice(c *cli.Context, amount float64, token string) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        CM: true,
        RPLExchange: true,
        LoadContracts: []string{"rocketPoolToken"},
        ClientConn: true,
        ClientSync: true,
        RocketStorage: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Get args
    amountWei := eth.EthToWei(amount)

    // Get token price
    price, err := exchange.GetTokenPrice(p, amountWei, token)
    if err != nil { return err }

    // Print response
    api.PrintResponse(p.Output, price)
    return nil

}

