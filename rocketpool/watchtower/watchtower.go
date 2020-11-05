package watchtower

import (
    "net/http"
    "time"

    "github.com/fatih/color"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/log"
)


// Config
var tasksInterval, _ = time.ParseDuration("5m")
var taskCooldown, _ = time.ParseDuration("1m")
const (
    MaxConcurrentEth1Requests = 200

    SubmitNetworkBalancesColor = color.FgYellow
    SubmitWithdrawableMinipoolsColor = color.FgBlue
    DissolveTimedOutMinipoolsColor = color.FgMagenta
    ProcessWithdrawalsColor = color.FgCyan
    ErrorColor = color.FgRed
)


// Register watchtower command
func RegisterCommands(app *cli.App, name string, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Run Rocket Pool watchtower activity daemon",
        Action: func(c *cli.Context) error {
            return run(c)
        },
    })
}


// Run daemon
func run(c *cli.Context) error {

    // Configure
    configureHTTP()

    // Wait until node is registered
    if err := services.WaitNodeRegistered(c, true); err != nil { return err }

    // Initialize tasks
    submitNetworkBalances, err := newSubmitNetworkBalances(c, log.NewColorLogger(SubmitNetworkBalancesColor))
    if err != nil { return err }
    submitWithdrawableMinipools, err := newSubmitWithdrawableMinipools(c, log.NewColorLogger(SubmitWithdrawableMinipoolsColor))
    if err != nil { return err }
    dissolveTimedOutMinipools, err := newDissolveTimedOutMinipools(c, log.NewColorLogger(DissolveTimedOutMinipoolsColor))
    if err != nil { return err }
    processWithdrawals, err := newProcessWithdrawals(c, log.NewColorLogger(ProcessWithdrawalsColor))
    if err != nil { return err }

    // Initialize error logger
    errorLog := log.NewColorLogger(ErrorColor)

    // Run task loop
    for {
        if err := submitNetworkBalances.run(); err != nil {
            errorLog.Println(err)
        }
        time.Sleep(taskCooldown)
        if err := submitWithdrawableMinipools.run(); err != nil {
            errorLog.Println(err)
        }
        time.Sleep(taskCooldown)
        if err := dissolveTimedOutMinipools.run(); err != nil {
            errorLog.Println(err)
        }
        time.Sleep(taskCooldown)
        if err := processWithdrawals.run(); err != nil {
            errorLog.Println(err)
        }
        time.Sleep(tasksInterval)
    }

}


// Configure HTTP transport settings
func configureHTTP() {

    // The watchtower daemon makes a large number of concurrent RPC requests to the Eth1 client
    // The HTTP transport is set to cache connections for future re-use equal to the maximum expected number of concurrent requests
    // This prevents issues related to memory consumption and address allowance from repeatedly opening and closing connections
    http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = MaxConcurrentEth1Requests

}

