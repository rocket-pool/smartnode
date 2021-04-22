package main

import (
	"log"
	"os"
	"sync"

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
    app.Version = "1.0.0-beta.4"
    app.Authors = []cli.Author{
        cli.Author{
            Name:  "David Rugendyke",
            Email: "david@rocketpool.net",
        },
        cli.Author{
            Name:  "Jake Pospischil",
            Email: "jake@rocketpool.net",
        },
        cli.Author{
            Name:  "Joe Clapis",
            Email: "joe@rocketpool.net",
        },
    }
    app.Copyright = "(c) 2021 Rocket Pool Pty Ltd"

    // Configure application
    app.Flags = []cli.Flag{
        cli.StringFlag{
            Name:  "httpPort, p",
            Usage: "Local HTTP port to listen on",
            Value: "8545",
        },
        cli.StringFlag{
            Name:  "wsPort, w",
            Usage: "Local Websocket port to listen on",
            Value: "8546",
        },
        cli.StringFlag{
            Name:  "httpProviderUrl, u",
            Usage: "External Eth 1.0 provider HTTP `URL`, including the remote port (defaults to Infura)",
            Value: "",
        },
        cli.StringFlag{
            Name:  "wsProviderUrl, r",
            Usage: "External Eth 1.0 provider Websocket `URL`, including the remote port (defaults to Infura)",
            Value: "",
        },
        cli.StringFlag{
            Name:  "network, n",
            Usage: "`Network` to connect to via Infura",
            Value: "goerli",
        },
        cli.StringFlag{
            Name:  "projectId, i",
            Usage: "Infura `project ID` to use for connection",
            Value: "",
        },
    }

    // Set application action
    app.Action = func(c *cli.Context) error {

        // We need a wait group since we have 2 HTTP listeners
        wg := new(sync.WaitGroup)
        wg.Add(2)

        // HTTP server
        go func() {
            proxyServer := proxy.NewHttpProxyServer(c.GlobalString("httpPort"), c.GlobalString("httpProviderUrl"), c.GlobalString("network"), c.GlobalString("projectId"))
            proxyServer.Start()
            wg.Done()
        }()
    
        // Websocket server
        go func() {
            proxyServer := proxy.NewWsProxyServer(c.GlobalString("wsPort"), c.GlobalString("wsProviderUrl"), c.GlobalString("network"), c.GlobalString("projectId"))
            proxyServer.Start()
            wg.Done()
        }()

        // Wait for both servers to stop
        wg.Wait()
		return nil
		
    }

    // Run application
    if err := app.Run(os.Args); err != nil {
        log.Fatal(err)
    }

}
