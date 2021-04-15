package main

import (
    "log"
    "os"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/rocketpool-pow-proxy/proxy"
)


// Run
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
    app.Copyright = "(c) 2020 Rocket Pool Pty Ltd"

    // Configure application
    app.Flags = []cli.Flag{
        cli.StringFlag{
			Name:  "port, p",
			Usage: "Port to listen on",
			Value: "8545",
		},
		cli.StringFlag{
			Name:  "providerUrl, u",
			Usage: "External Eth 1.0 provider `URL` (defaults to Infura)",
			Value: "",
		},
		cli.StringFlag{
			Name:  "providerType, t",
			Usage: "Eth 1.0 provider type if not using `URL`: Infura or Pocket",
			Value: "infura",
		},
		cli.StringFlag{
			Name:  "network, n",
			Usage: "`Network` to connect to via Infura / Pocket",
			Value: "goerli",
		},
		cli.StringFlag{
			Name:  "projectId, i",
			Usage: "Infura / Pocket `project ID` to use for connection; for Pocket load balancers prefix with lb/",
			Value: "",
		},
    }

    // Set application action
    app.Action = func(c *cli.Context) error {

        // Initialise and start proxy server
		proxyServer := proxy.NewProxyServer(c.GlobalString("port"), c.GlobalString("providerUrl"), c.GlobalString("providerType"), c.GlobalString("network"), c.GlobalString("projectId"))
        return proxyServer.Start()

    }

    // Run application
    if err := app.Run(os.Args); err != nil {
        log.Fatal(err)
    }

}

