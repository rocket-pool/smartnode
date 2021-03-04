package node

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
var taskCooldown, _ = time.ParseDuration("10s")
const (
    MaxConcurrentEth1Requests = 200

    ClaimRplRewardsColor = color.FgGreen
    StakePrelaunchMinipoolsColor = color.FgBlue
    ErrorColor = color.FgRed
)


// Register node command
func RegisterCommands(app *cli.App, name string, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Run Rocket Pool node activity daemon",
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
    claimRplRewards, err := newClaimRplRewards(c, log.NewColorLogger(ClaimRplRewardsColor))
    if err != nil { return err }
    stakePrelaunchMinipools, err := newStakePrelaunchMinipools(c, log.NewColorLogger(StakePrelaunchMinipoolsColor))
    if err != nil { return err }

    // Initialize error logger
    errorLog := log.NewColorLogger(ErrorColor)

    // Run task loop
    for {
        if err := claimRplRewards.run(); err != nil {
            errorLog.Println(err)
        }
        time.Sleep(taskCooldown)
        if err := stakePrelaunchMinipools.run(); err != nil {
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

