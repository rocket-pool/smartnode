package watchtower

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
)


// Register watchtower command
func RegisterCommands(app *cli.App, name string, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Run Rocket Pool watchtower daemon",
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
        Client: true,
        CM: true,
        Beacon: true,
        LoadContracts: []string{"rocketMinipoolSettings", "rocketNodeAPI", "rocketNodeWatchtower", "rocketPool"},
        LoadAbis: []string{"rocketMinipool"},
        WaitPassword: true,
        WaitNodeAccount: true,
        WaitClientConn: true,
        WaitClientSync: true,
        WaitRocketStorage: true,
    })
    if err != nil { return err }

    // Start minipool processes
    go StartWatchtowerProcess(p)

    // Block thread
    select {}

}

