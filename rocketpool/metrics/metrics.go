package metrics

import (
    "net/http"
    "time"

    "github.com/fatih/color"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/utils/log"
)


// Config
const (
    maxConcurrentEth1Requests = 200
    networkMetricsColor = color.BgYellow
    minipoolMetricsColor = color.BgGreen
    nodeMetricsColor = color.BgCyan
    errorColor = color.FgRed
)
var networkUpdateInterval, _ = time.ParseDuration("30s")
var minipoolUpdateInterval, _ = time.ParseDuration("30s")
var nodeUpdateInterval, _ = time.ParseDuration("5m")


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
    logger := log.NewColorLogger(networkMetricsColor)
    errorLog := log.NewColorLogger(errorColor)
    logger.Println("Enter metrics.run")

    // Configure
    configureHTTP()

    // Start metrics processes
    go (func() { startNetworkMetricsProcess(c, networkUpdateInterval, log.NewColorLogger(networkMetricsColor)) })()
    go (func() { startMinipoolMetricsProcess(c, minipoolUpdateInterval, log.NewColorLogger(minipoolMetricsColor)) })()
    go (func() { startNodeMetricsProcess(c, nodeUpdateInterval, log.NewColorLogger(nodeMetricsColor)) })()

    // Serve metrics
    http.Handle("/metrics", promhttp.Handler())
    err := http.ListenAndServe(":2112", nil)
    if (err != nil) {
        errorLog.Printlnf("Exit metrics.run with error: %w", err)
    } else {
        logger.Println("Exit metrics.run")
    }

    return err
}


// Configure HTTP transport settings
func configureHTTP() {

    // The watchtower daemon makes a large number of concurrent RPC requests to the Eth1 client
    // The HTTP transport is set to cache connections for future re-use equal to the maximum expected number of concurrent requests
    // This prevents issues related to memory consumption and address allowance from repeatedly opening and closing connections
    http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = maxConcurrentEth1Requests

}
