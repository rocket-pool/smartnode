package exchange

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/api/exchange"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/api"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Buy tokens with ether
func buyTokens(c *cli.Context, etherAmount float64, tokenAmount float64, token string) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        Client: true,
        CM: true,
        RPLExchangeAddress: true,
        LoadContracts: []string{"rocketPoolToken"},
        ClientConn: true,
        ClientSync: true,
        RocketStorage: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Get args
    etherAmountWei := eth.EthToWei(etherAmount)
    tokenAmountWei := eth.EthToWei(tokenAmount)

    // Check tokens can be bought
    canBuy, err := exchange.CanBuyTokens(p, etherAmountWei, tokenAmountWei, token)
    if err != nil { return err }

    // Check response
    if !canBuy.Success {
        api.PrintResponse(p.Output, canBuy)
        return nil
    }

    // Buy tokens
    bought, err := exchange.BuyTokens(p, etherAmountWei, tokenAmountWei, token)
    if err != nil { return err }

    // Print response
    api.PrintResponse(p.Output, bought)
    return nil

}

