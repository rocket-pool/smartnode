package main

import (
    "fmt"
    "log"
    "os"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/commands/deposits"
    "github.com/rocket-pool/smartnode-cli/rocketpool/commands/fees"
    "github.com/rocket-pool/smartnode-cli/rocketpool/commands/node"
    "github.com/rocket-pool/smartnode-cli/rocketpool/commands/resources"
    "github.com/rocket-pool/smartnode-cli/rocketpool/commands/rpip"
    "github.com/rocket-pool/smartnode-cli/rocketpool/commands/users"
)


// Run
func main() {

    // Add logo to application help template
    cli.AppHelpTemplate = fmt.Sprintf(
        "______           _        _    ______           _ " + "\n" +
        "| ___ \\         | |      | |   | ___ \\         | |" + "\n" +
        "| |_/ /___   ___| | _____| |_  | |_/ /__   ___ | |" + "\n" +
        "|    // _ \\ / __| |/ / _ \\ __| |  __/ _ \\ / _ \\| |" + "\n" +
        "| |\\ \\ (_) | (__|   <  __/ |_  | | | (_) | (_) | |" + "\n" +
        "\\_| \\_\\___/ \\___|_|\\_\\___|\\__| \\_|  \\___/ \\___/|_|" + "\n\n" +
    "%s", cli.AppHelpTemplate)

    // Initialise application
    app := cli.NewApp()

    // Configure application
    app.Name     = "rocketpool"
    app.Usage    = "Rocket Pool node operator utilities"
    app.Version  = "0.0.1"
    app.Authors  = []cli.Author{
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
    app.Copyright = "(c) 2018 Rocket Pool Pty Ltd"

    // Register commands
    deposits.RegisterCommands(app, []string{"d"})
    fees.RegisterCommands(app, []string{"f"})
    node.RegisterCommands(app, []string{"n"})
    resources.RegisterCommands(app, []string{"r"})
    rpip.RegisterCommands(app, nil)
    users.RegisterCommands(app, []string{"u"})

    // Run application
    err := app.Run(os.Args)
    if err != nil {
        log.Fatal(err)
    }

}

