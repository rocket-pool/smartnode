package minipools

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
)


// Register minipools command
func RegisterCommands(app *cli.App, name string, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Run Rocket Pool minipool management daemon",
        Action: func(c *cli.Context) error {
            return run(c)
        },
    })
}


// Run process
func run(c *cli.Context) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        KM: true,
        Client: true,
        CM: true,
        NodeContractAddress: true,
        Beacon: true,
        Docker: true,
        LoadContracts: []string{"utilAddressSetStorage", "rocketNodeAPI"},
        LoadAbis: []string{"rocketMinipool", "rocketNodeContract"},
        WaitPassword: true,
        WaitNodeAccount: true,
        WaitNodeRegistered: true,
        WaitClientConn: true,
        WaitClientSync: true,
        WaitRocketStorage: true,
    })
    if err != nil { return err }

    // Start minipools management process
    go StartMinipoolsProcess(p)

    // Block thread
    select {}

}

