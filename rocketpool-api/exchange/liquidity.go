package exchange

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/api/exchange"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/api"
)


// Get the current token liquidity
func getTokenLiquidity(c *cli.Context, token string) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        CM: true,
        RPLExchangeAddress: true,
        LoadContracts: []string{"rocketPoolToken"},
        ClientConn: true,
        ClientSync: true,
        RocketStorage: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Get token liquidity
    liquidity, err := exchange.GetTokenLiquidity(p, token)
    if err != nil { return err }

    // Print response
    api.PrintResponse(p.Output, liquidity)
    return nil

}

