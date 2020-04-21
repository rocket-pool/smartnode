package deposit

import (
    "math/big"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/api/deposit"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/api"
)


// Can complete the reserved node deposit
func canCompleteDeposit(c *cli.Context) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
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

    // RPL send not available via API
    if canComplete.RplRequiredWei.Cmp(big.NewInt(0)) > 0 {
        canComplete.InsufficientNodeRplBalance = true
        canComplete.Success = false
    }

    // Get error message
    var message string
    if canComplete.ReservationDidNotExist {
        message = "Node does not have an existing deposit reservation"
    } else if canComplete.DepositsDisabled {
        message = "Node deposits are currently disabled in Rocket Pool"
    } else if canComplete.MinipoolCreationDisabled {
        message = "Minipool creation is currently disabled in Rocket Pool"
    } else if canComplete.InsufficientNodeEtherBalance {
        message = "Node has insufficient ETH balance to complete deposit"
    } else if canComplete.InsufficientNodeRplBalance {
        message = "Node has insufficient RPL balance to complete deposit"
    }

    // Return
    api.PrintResponse(p.Output, canComplete, message)
    return nil

}


// Complete the reserved node deposit
func completeDeposit(c *cli.Context) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
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

    // RPL send not available via API
    if canComplete.RplRequiredWei.Cmp(big.NewInt(0)) > 0 {
        canComplete.InsufficientNodeRplBalance = true
        canComplete.Success = false
    }

    // Check response
    if !canComplete.Success {
        var message string
        if canComplete.ReservationDidNotExist {
            message = "Node does not have an existing deposit reservation"
        } else if canComplete.DepositsDisabled {
            message = "Node deposits are currently disabled in Rocket Pool"
        } else if canComplete.MinipoolCreationDisabled {
            message = "Minipool creation is currently disabled in Rocket Pool"
        } else if canComplete.InsufficientNodeEtherBalance {
            message = "Node has insufficient ETH balance to complete deposit"
        } else if canComplete.InsufficientNodeRplBalance {
            message = "Node has insufficient RPL balance to complete deposit"
        }
        api.PrintResponse(p.Output, canComplete, message)
        return nil
    }

    // Complete deposit reservation
    completed, err := deposit.CompleteDeposit(p, canComplete.EtherRequiredWei, canComplete.DepositDurationId)
    if err != nil { return err }

    // Print response
    api.PrintResponse(p.Output, completed, "")
    return nil

}

