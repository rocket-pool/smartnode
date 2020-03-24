package deposit

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/api/deposit"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/api"
)


// Reserve a node deposit
func reserveDeposit(c *cli.Context, durationId string) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        Client: true,
        CM: true,
        NodeContractAddress: true,
        NodeContract: true,
        LoadContracts: []string{"rocketNodeAPI", "rocketMinipoolSettings", "rocketNodeSettings"},
        LoadAbis: []string{"rocketNodeContract"},
        ClientConn: true,
        ClientSync: true,
        RocketStorage: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Check node deposit can be reserved
    canReserve, err := deposit.CanReserveDeposit(p, durationId)
    if err != nil { return err }

    // Check response
    if !canReserve.Success {
        var message string
        if canReserve.HadExistingReservation {
            message = "Node has an existing deposit reservation"
        } else if canReserve.DepositsDisabled {
            message = "Node deposits are currently disabled in Rocket Pool"
        } else if canReserve.StakingDurationDisabled {
            message = "The specified staking duration is invalid or disabled"
        }
        api.PrintResponse(p.Output, canReserve, message)
        return nil
    }

    // Reserve node deposit
    reserved, err := deposit.ReserveDeposit(p, durationId)
    if err != nil { return err }

    // Print response
    api.PrintResponse(p.Output, reserved, "")
    return nil

}

