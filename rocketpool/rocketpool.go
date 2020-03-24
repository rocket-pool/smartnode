package main

import (
    "log"
    "os"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/rocketpool/api"
    rpcli "github.com/rocket-pool/smartnode/rocketpool/cli"
    //"github.com/rocket-pool/smartnode/rocketpool/minipool"
    //"github.com/rocket-pool/smartnode/rocketpool/minipools"
    //"github.com/rocket-pool/smartnode/rocketpool/node"
    //proxy "github.com/rocket-pool/smartnode/rocketpool/pow-proxy"
    //"github.com/rocket-pool/smartnode/rocketpool/watchtower"
    apiutils "github.com/rocket-pool/smartnode/shared/utils/api"
    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


// Run
func main() {

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

    // Configure application
    cliutils.Configure(app)

    // Register commands
           api.RegisterCommands(app, "api",        []string{"a"})
         rpcli.RegisterCommands(app, "run",        []string{"r"})
    //  minipool.RegisterCommands(app, "minipool",   []string{"m"})
    // minipools.RegisterCommands(app, "minipools",  []string{"p"})
    //      node.RegisterCommands(app, "node",       []string{"n"})
    //     proxy.RegisterCommands(app, "pow-proxy",  []string{"x"})
    //watchtower.RegisterCommands(app, "watchtower", []string{"w"})

    // Run application
    if err := app.Run(os.Args); err != nil {
        if len(os.Args) > 1 && os.Args[1] == "api" {
            apiutils.PrintErrorResponse(nil, err)
        } else {
            log.Fatal(err)
        }
    }

}

