package main

import (
    "log"
    "os"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/rocketpool-pow-proxy/proxy"
)



// Run application
func main() {

    // Initialise application
    app := cli.NewApp()

    // Set application info
    app.Name = "rocketpool-pow-proxy"
    app.Usage = "Rocket Pool Eth 1.0 proxy server"
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
    app.Copyright = "(c) 2019 Rocket Pool Pty Ltd"

    // Configure application
    app.Flags = []cli.Flag{
        cli.StringFlag{
            Name:  "port",
            Usage: "Port to listen on",
            Value: "8545",
        },
        cli.StringFlag{
            Name:  "providerUrl",
            Usage: "External Eth 1.0 provider `URL` (defaults to Infura)",
            Value: "",
        },
        cli.StringFlag{
            Name:  "network",
            Usage: "`Network` to connect to via Infura",
            Value: "goerli",
        },
        cli.StringFlag{
            Name:  "projectId",
            Usage: "Infura `project ID` to use for connection",
            Value: "",
        },
    };

    // Set application action
    app.Action = func(c *cli.Context) error {

        // Initialise and start proxy server
        proxyServer := proxy.NewProxyServer(c.GlobalString("port"), c.GlobalString("providerUrl"), c.GlobalString("network"), c.GlobalString("projectId"))
        return proxyServer.Start()

    }

    // Run application
    if err := app.Run(os.Args); err != nil {
        log.Fatal(err)
    }

}

