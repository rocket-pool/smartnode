package main

import (
	"log"
	"os"
	"sync"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/rocketpool-pow-proxy/proxy"
	"github.com/rocket-pool/smartnode/shared"
)

// Run
func main() {

    // Initialise application
    app := cli.NewApp()

    // Set application info
    app.Name = "rocketpool-pow-proxy"
    app.Usage = "Rocket Pool Eth 1.0 proxy server"
    app.Version = shared.RocketPoolVersion
    app.Authors = []cli.Author{
        {
            Name:  "David Rugendyke",
            Email: "david@rocketpool.net",
        },
        {
            Name:  "Jake Pospischil",
            Email: "jake@rocketpool.net",
        },
        {
            Name:  "Joe Clapis",
            Email: "joe@rocketpool.net",
        },
        {
            Name:  "Kane Wallmann",
            Email: "kane@rocketpool.net",
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
            Usage: "External Eth 1.0 provider HTTP `URL`, including the remote port (ignored if 'providerType' is used)",
            Value: "",
        },
        cli.StringFlag{
            Name:  "wsProviderUrl, r",
            Usage: "External Eth 1.0 provider Websocket `URL`, including the remote port (ignored if 'providerType' is used)",
            Value: "",
        },
        cli.StringFlag{
            Name:  "network, n",
            Usage: "`Network` to connect to via Infura",
            Value: "goerli",
        },
        cli.StringFlag{
            Name:  "projectId, i",
            Usage: "Infura project ID or Pocket App ID to use for connection; for Pocket load balancers, prefix with \"lb/\"",
            Value: "",
        },
        cli.StringFlag{
            Name:  "providerType, t",
            Usage: "Eth 1.0 provider type if not using `URL`: Infura or Pocket",
            Value: "infura",
        },
        cli.BoolFlag{
            Name:  "verbose, V",
            Usage: "Enables logging of all incoming and outgoing proxied data",
        },
    }

    // Set application action
    app.Action = func(c *cli.Context) error {

        // We need a wait group since we have 2 HTTP listeners
        wg := new(sync.WaitGroup)
        wg.Add(2)

        // HTTP server
        go func() {
            proxyServer := proxy.NewHttpProxyServer(c.GlobalString("httpPort"), c.GlobalString("httpProviderUrl"), c.GlobalString("network"), c.GlobalString("projectId"), c.GlobalString("providerType"), c.GlobalBool("verbose"))
            err := proxyServer.Start()
            if err != nil {
                log.Fatalf("Could not start HTTP proxy server %v", err)
                return
            }
            wg.Done()
        }()
    
        // Websocket server
        go func() {
            if c.GlobalString("providerType") == "infura" || c.GlobalString("wsProviderUrl") != "" {
                proxyServer := proxy.NewWsProxyServer(c.GlobalString("wsPort"), c.GlobalString("wsProviderUrl"), c.GlobalString("network"), c.GlobalString("projectId"), c.GlobalBool("verbose"))
                err := proxyServer.Start()
                if err != nil {
                    log.Fatalf("Could not start websocket proxy server %v", err)
                    return
                }
            } else {
                log.Println("No websocket URL provided, running in HTTP-only mode.")
            }
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
