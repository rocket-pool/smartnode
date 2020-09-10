package watchtower

import (
    "github.com/fatih/color"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/log"
)


// Config
const (
    DissolveTimedOutMinipoolsColor = color.FgMagenta
    ProcessWithdrawalsColor = color.FgCyan
    SubmitNetworkBalancesColor = color.FgYellow
    SubmitWithdrawableMinipoolsColor = color.FgBlue
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

    // Wait until node is registered
    if err := services.WaitNodeRegistered(c, true); err != nil { return err }

    // Initialize tasks
    dissolveTimedOutMinipools, err := newDissolveTimedOutMinipools(c, log.NewColorLogger(DissolveTimedOutMinipoolsColor))
    if err != nil { return err }
    processWithdrawals, err := newProcessWithdrawals(c, log.NewColorLogger(ProcessWithdrawalsColor))
    if err != nil { return err }
    submitNetworkBalances, err := newSubmitNetworkBalances(c, log.NewColorLogger(SubmitNetworkBalancesColor))
    if err != nil { return err }
    submitWithdrawableMinipools, err := newSubmitWithdrawableMinipools(c, log.NewColorLogger(SubmitWithdrawableMinipoolsColor))
    if err != nil { return err }

    // Start tasks
    dissolveTimedOutMinipools.Start()
    processWithdrawals.Start()
    submitNetworkBalances.Start()
    submitWithdrawableMinipools.Start()

    // Block thread
    select {}

}

