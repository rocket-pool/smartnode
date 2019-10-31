package deposit

import (
    "math/big"

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
    canComplete, err := deposit.CanCompleteDeposit(p)
    if err != nil { return err }

    // RPL send not available
    if canComplete.RplRequiredWei.Cmp(big.NewInt(0)) > 0 {
        canComplete.InsufficientNodeRplBalance = true
    }

    // Check response
    if canComplete.ReservationDidNotExist || canComplete.DepositsDisabled || canComplete.MinipoolCreationDisabled || canComplete.InsufficientNodeEtherBalance || canComplete.InsufficientNodeRplBalance {
        api.PrintResponse(p.Output, canComplete)
        return nil
    }

    // Complete deposit reservation
    completed, err := deposit.CompleteDeposit(p, canComplete.EtherRequiredWei, canComplete.DepositDurationId)
    if err != nil { return err }

    // Print response
    api.PrintResponse(p.Output, completed)
    return nil

}

