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

            // Check argument count
            if len(c.Args()) != 4 {
                return cli.NewExitError("USAGE:" + "\n   " + "rocketpool minipools rpPath imageName containerPrefix rpNetwork", 1)
            }

            // Get arguments
            rpPath := c.Args().Get(0)
            imageName := c.Args().Get(1)
            containerPrefix := c.Args().Get(2)
            rpNetwork := c.Args().Get(3)

            // Run process
            return run(c, rpPath, imageName, containerPrefix, rpNetwork)

        },
    })
}


// Run process
func run(c *cli.Context, rpPath string, imageName string, containerPrefix string, rpNetwork string) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        CM: true,
        Docker: true,
        LoadContracts: []string{"utilAddressSetStorage"},
        LoadAbis: []string{"rocketMinipool"},
        WaitPassword: true,
        WaitNodeAccount: true,
        WaitClientConn: true,
        WaitClientSync: true,
        WaitRocketStorage: true,
    })
    if err != nil { return err }

    // Start minipools management process
    go StartManagementProcess(p, rpPath, imageName, containerPrefix, rpNetwork)

    // Block thread
    select {}

}

