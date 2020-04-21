package metrics

import (
    "net/http"
    "time"

    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/urfave/cli"
)


// Config
const UPDATE_METRICS_INTERVAL string = "15s"
var updateMetricsInterval, _ = time.ParseDuration(UPDATE_METRICS_INTERVAL)


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

    // Start metrics processes
    go StartEth1MetricsProcess(c)
    go StartEth2MetricsProcess(c)
    go StartRocketPoolMetricsProcess(c)

    // Serve metrics
    http.Handle("/metrics", promhttp.Handler())
    return http.ListenAndServe(":2112", nil)

}

