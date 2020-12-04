package main

import (
    "fmt"
    "os"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/rocketpool-cli/minipool"
    "github.com/rocket-pool/smartnode/rocketpool-cli/network"
    "github.com/rocket-pool/smartnode/rocketpool-cli/node"
    "github.com/rocket-pool/smartnode/rocketpool-cli/queue"
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
    app.Version = "0.0.9"
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
        cli.BoolFlag{
            Name:  "allow-root, r",
            Usage: "Allow rocketpool to be run as the root user",
        },
        cli.StringFlag{
            Name:  "config-path, c",
            Usage: "Rocket Pool config asset `path`",
            Value: "~/.rocketpool",
        },
        cli.StringFlag{
            Name:  "daemon-path, d",
            Usage: "Interact with a Rocket Pool service daemon at a `path` on the host OS, running outside of docker",
        },
        cli.StringFlag{
            Name:  "host, o",
            Usage: "Smart node SSH host `address`",
        },
        cli.StringFlag{
            Name:  "user, u",
            Usage: "Smart node SSH user `name`",
        },
        cli.StringFlag{
            Name:  "key, k",
            Usage: "Smart node SSH key `file`",
        },
        cli.StringFlag{
            Name:  "passphrase, p",
            Usage: "Smart node SSH key `passphrase`",
        },
    }

    // Register commands
    minipool.RegisterCommands(app, "minipool", []string{"m"})
     network.RegisterCommands(app, "network",  []string{"e"})
        node.RegisterCommands(app, "node",     []string{"n"})
       queue.RegisterCommands(app, "queue",    []string{"q"})
     service.RegisterCommands(app, "service",  []string{"s"})
      wallet.RegisterCommands(app, "wallet",   []string{"w"})

    // Check user ID
    app.Before = func(c *cli.Context) error {
        if os.Getuid() == 0 && !c.GlobalBool("allow-root") {
            fmt.Fprintln(os.Stderr, "rocketpool should not be run as root. Please try again without 'sudo'.")
            fmt.Fprintln(os.Stderr, "If you want to run rocketpool as root anyway, use the '--allow-root' option to override this warning.")
            os.Exit(1)
        }
        return nil
    }

    // Run application
    fmt.Println("")
    if err := app.Run(os.Args); err != nil {
        fmt.Println(err)
    }
    fmt.Println("")

}

