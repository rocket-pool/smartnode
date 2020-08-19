package main

import (
    "fmt"
    "os"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/rocketpool-cli/config"
    "github.com/rocket-pool/smartnode/rocketpool-cli/network"
    "github.com/rocket-pool/smartnode/rocketpool-cli/service"
    "github.com/rocket-pool/smartnode/rocketpool-cli/wallet"
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
    app.Name = "rocketpool"
    app.Usage = "Rocket Pool CLI"
    app.Version = "0.0.1"
    app.Authors = []cli.Author{
        cli.Author{
            Name:  "David Rugendyke",
            Email: "david@rocketpool.net",
        },
        cli.Author{
            Name:  "Jake Pospischil",
            Email: "jake@rocketpool.net",
        },
    }
    app.Copyright = "(c) 2020 Rocket Pool Pty Ltd"

    // Set application flags
    app.Flags = []cli.Flag{
        cli.StringFlag{
            Name:  "host",
            Usage: "Smart node SSH host address",
        },
        cli.StringFlag{
            Name:  "user",
            Usage: "Smart node SSH user account",
        },
        cli.StringFlag{
            Name:  "key",
            Usage: "Smart node SSH key file",
        },
    }

    // Register commands
     config.RegisterCommands(app, "config",  []string{"c"})
    network.RegisterCommands(app, "network", []string{"n"})
    service.RegisterCommands(app, "service", []string{"s"})
     wallet.RegisterCommands(app, "wallet",  []string{"w"})

    // Run application
    fmt.Println("")
    if err := app.Run(os.Args); err != nil {
        fmt.Println(err)
    }
    fmt.Println("")

}

