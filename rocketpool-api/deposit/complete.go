package deposit

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/api/deposit"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/api"
)


// Complete the reserved node deposit
func completeDeposit(c *cli.Context) error {

	// Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
    	AM: true,
        KM: true,
        Client: true,
        CM: true,
        NodeContractAddress: true,
        NodeContract: true,
        LoadContracts: []string{"rocketDepositQueue", "rocketETHToken", "rocketMinipoolSettings", "rocketNodeAPI", "rocketNodeSettings", "rocketPool", "rocketPoolToken"},
        LoadAbis: []string{"rocketNodeContract"},
        ClientConn: true,
        ClientSync: true,
        RocketStorage: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Check deposit reservation can be completed
    response, err := deposit.CanCompleteDeposit(p)
    if err != nil { return err }

    // Check response
    if response.ReservationDidNotExist || response.DepositsDisabled || response.MinipoolCreationDisabled || response.InsufficientNodeEtherBalance || response.InsufficientNodeRplBalance {
        api.PrintResponse(p.Output, response)
        return nil
    }

    // Complete deposit reservation
    response, err = deposit.CompleteDeposit(p, response.TxValueWei)
    if err != nil { return err }

    // Print response
    api.PrintResponse(p.Output, response)
    return nil

}

