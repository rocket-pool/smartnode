package minipool

import (
    "github.com/ethereum/go-ethereum/common"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
)


// Register minipool command
func RegisterCommands(app *cli.App, name string, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Run Rocket Pool minipool activity daemon",
        Action: func(c *cli.Context) error {

            // Check argument count
            if len(c.Args()) != 1 {
                return cli.NewExitError("USAGE:" + "\n   " + "rocketpool minipool address", 1)
            }

            // Get & validate arguments
            address := c.Args().Get(0)
            if !common.IsHexAddress(address) {
                return cli.NewExitError("Invalid minipool address", 1)
            }

            // Run process
            return run(c, address)

        },
    })
}


// Run process
func run(c *cli.Context, address string) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        KM: true,
        Client: true,
        CM: true,
        NodeContractAddress: true,
        Beacon: true,
        Docker: true,
        LoadContracts: []string{"rocketNodeAPI"},
        LoadAbis: []string{"rocketMinipool", "rocketNodeContract"},
        WaitPassword: true,
        WaitNodeAccount: true,
        WaitNodeRegistered: true,
        WaitClientConn: true,
        WaitClientSync: true,
        WaitRocketStorage: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Initialise minipool
    pool, err := Initialise(p, address)
    if err != nil { return err }

    // Stake minipool
    if err := Stake(p, pool); err != nil { return err }

    // Process done channel
    done := make(chan struct{})

    // Start minipool processes
    go StartWithdrawalProcess(p, pool, done)

    // Block thread until done
    for received := 0; received < 2; {
        select {
            case <-done:
                received++
        }
    }
    return nil

}

