package deposit

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/api/deposit"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/api"
)


// Cancel the reserved node deposit
func cancelDeposit(c *cli.Context) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        Client: true,
        CM: true,
        NodeContractAddress: true,
        NodeContract: true,
        LoadContracts: []string{"rocketNodeAPI"},
        LoadAbis: []string{"rocketNodeContract"},
        ClientConn: true,
        ClientSync: true,
        RocketStorage: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Check deposit reservation can be cancelled
    response, err := deposit.CanCancelDeposit(p)
    if err != nil { return err }

    // Check response
    if response.ReservationDidNotExist {
        api.PrintResponse(p.Output, response)
        return nil
    }

    // Cancel deposit reservation
    response, err = deposit.CancelDeposit(p)
    if err != nil { return err }

    // Print response
    api.PrintResponse(p.Output, response)
    return nil

}

