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
        LoadContracts: []string{"rocketNodeAPI", "rocketNodeSettings"},
        LoadAbis: []string{"rocketNodeContract"},
        ClientConn: true,
        ClientSync: true,
        RocketStorage: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Reserve deposit & print response
    if response, err := deposit.ReserveDeposit(p, durationId); err != nil {
        return err
    } else {
        api.PrintResponse(p.Output, response)
        return nil
    }

}

