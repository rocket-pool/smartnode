package metrics

import (
    "net/http"

    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
)


// Register metrics command
func RegisterCommands(app *cli.App, name string, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Run Rocket Pool metrics daemon",
        Action: func(c *cli.Context) error {
            return run(c)
        },
    })
}


// Run process
func run(c *cli.Context) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        Client: true,
        WaitClientConn: true,
    })
    if err != nil { return err }

    // Register metrics
    registerEth1Metrics(p)

    // Serve metrics
    http.Handle("/metrics", promhttp.Handler())
    return http.ListenAndServe(":2112", nil)

}

