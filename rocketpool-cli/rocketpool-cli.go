package main

import (
    "fmt"
    "log"
    "os"

    "gopkg.in/urfave/cli.v1"

    "github.com/rocket-pool/smartnode-cli/rocketpool-cli/deposit"
    "github.com/rocket-pool/smartnode-cli/rocketpool-cli/fee"
    "github.com/rocket-pool/smartnode-cli/rocketpool-cli/minipool"
    "github.com/rocket-pool/smartnode-cli/rocketpool-cli/node"
    cliutils "github.com/rocket-pool/smartnode-cli/shared/utils/cli"
)

// Run
func main() {

    // Add logo to application help template
    cli.AppHelpTemplate = fmt.Sprintf(`
______           _        _    ______           _ 
| ___ \         | |      | |   | ___ \         | |
| |_/ /___   ___| | _____| |_  | |_/ /__   ___ | |
|    // _ \ / __| |/ / _ \ __| |  __/ _ \ / _ \| |
| |\ \ (_) | (__|   <  __/ |_  | | | (_) | (_) | |
\_| \_\___/ \___|_|\_\___|\__| \_|  \___/ \___/|_|

%s`, cli.AppHelpTemplate)

    // Initialise application
    app := cli.NewApp()

    // Set application info
    app.Name = "rocketpool-cli"
    app.Usage = "Rocket Pool node operator utilities"
    app.Version = "0.0.1"
    app.Authors = []cli.Author{
        cli.Author{
            Name:  "Darren Langley",
            Email: "darren@rocketpool.net",
        },
        cli.Author{
            Name:  "David Rugendyke",
            Email: "david@rocketpool.net",
        },
        cli.Author{
            Name:  "Jake Pospischil",
            Email: "jake@rocketpool.net",
        },
    }
    app.Copyright = "(c) 2019 Rocket Pool Pty Ltd"

    // Configure application
    cliutils.Configure(app)

    // Register commands
    deposit.RegisterCommands(app, "deposit", []string{"d"})
    fee.RegisterCommands(app, "fee", []string{"f"})
    minipool.RegisterCommands(app, "minipool", []string{"m"})
    node.RegisterCommands(app, "node", []string{"n"})

    // Run application
    if err := app.Run(os.Args); err != nil {
        log.Fatal(err)
    }

}
