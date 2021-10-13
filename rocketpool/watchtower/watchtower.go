package watchtower

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/fatih/color"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

// Config
var minTasksInterval, _ = time.ParseDuration("4m")
var maxTasksInterval, _ = time.ParseDuration("6m")
var taskCooldown, _ = time.ParseDuration("10s")
const (
    MaxConcurrentEth1Requests = 200

    RespondChallengesColor = color.FgWhite
    ClaimRplRewardsColor = color.FgGreen
    SubmitRplPriceColor = color.FgYellow
    SubmitNetworkBalancesColor = color.FgYellow
    SubmitWithdrawableMinipoolsColor = color.FgBlue
    DissolveTimedOutMinipoolsColor = color.FgMagenta
    ProcessWithdrawalsColor = color.FgCyan
    SubmitScrubMinipoolsColor = color.FgHiGreen
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
    respondChallenges, err := newRespondChallenges(c, log.NewColorLogger(RespondChallengesColor))
    if err != nil { return err }
    claimRplRewards, err := newClaimRplRewards(c, log.NewColorLogger(ClaimRplRewardsColor))
    if err != nil { return err }
    submitRplPrice, err := newSubmitRplPrice(c, log.NewColorLogger(SubmitRplPriceColor))
    if err != nil { return err }
    submitNetworkBalances, err := newSubmitNetworkBalances(c, log.NewColorLogger(SubmitNetworkBalancesColor))
    if err != nil { return err }
    submitWithdrawableMinipools, err := newSubmitWithdrawableMinipools(c, log.NewColorLogger(SubmitWithdrawableMinipoolsColor))
    if err != nil { return err }
    dissolveTimedOutMinipools, err := newDissolveTimedOutMinipools(c, log.NewColorLogger(DissolveTimedOutMinipoolsColor))
    if err != nil { return err }
    processWithdrawals, err := newProcessWithdrawals(c, log.NewColorLogger(ProcessWithdrawalsColor))
    if err != nil { return err }
    submitScrubMinipools, err := newSubmitScrubMinipools(c, log.NewColorLogger(SubmitScrubMinipoolsColor))
    if err != nil { return err }

    // Initialize error logger
    errorLog := log.NewColorLogger(ErrorColor)

    intervalDelta := maxTasksInterval - minTasksInterval
    secondsDelta := intervalDelta.Seconds()

    // Run task loop
    for {

        // Randomize the next interval
        randomSeconds := rand.Intn(int(secondsDelta))
        interval := time.Duration(randomSeconds) * time.Second + minTasksInterval

        if err := respondChallenges.run(); err != nil {
            errorLog.Println(err)
        }
        time.Sleep(taskCooldown)
        if err := claimRplRewards.run(); err != nil {
            errorLog.Println(err)
        }
        time.Sleep(taskCooldown)
        if err := submitRplPrice.run(); err != nil {
            errorLog.Println(err)
        }
        time.Sleep(taskCooldown)
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
        time.Sleep(taskCooldown)
        if err := submitScrubMinipools.run(); err != nil {
            errorLog.Println(err)
        }
        time.Sleep(interval)
    }

}


// Configure HTTP transport settings
func configureHTTP() {

    // The watchtower daemon makes a large number of concurrent RPC requests to the Eth1 client
    // The HTTP transport is set to cache connections for future re-use equal to the maximum expected number of concurrent requests
    // This prevents issues related to memory consumption and address allowance from repeatedly opening and closing connections
    http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = MaxConcurrentEth1Requests

}

