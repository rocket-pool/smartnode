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
        KM: true,
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

    // Generate new validator key
    validatorKey, err := p.KM.CreateValidatorKey()
    if err != nil { return err }

    // Check node deposit can be reserved
    canReserve, err := deposit.CanReserveDeposit(p, validatorKey, durationId)
    if err != nil { return err }

    // Check response
    if !canReserve.Success {
        api.PrintResponse(p.Output, canReserve)
        return nil
    }

    // Reserve node deposit
    reserved, err := deposit.ReserveDeposit(p, validatorKey, durationId)
    if err != nil { return err }

    // Print response
    api.PrintResponse(p.Output, reserved)
    return nil

}

