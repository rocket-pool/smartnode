package watchtower

import (
    "github.com/urfave/cli"
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

    // Initialize tasks
    dissolveTimedOutMinipools, err := newDissolveTimedOutMinipools(c)
    if err != nil { return err }
    processWithdrawals, err := newProcessWithdrawals(c)
    if err != nil { return err }
    submitNetworkBalances, err := newSubmitNetworkBalances(c)
    if err != nil { return err }
    submitWithdrawableMinipools, err := newSubmitWithdrawableMinipools(c)
    if err != nil { return err }

    // Start tasks
    dissolveTimedOutMinipools.Start()
    processWithdrawals.Start()
    submitNetworkBalances.Start()
    submitWithdrawableMinipools.Start()

    // Block thread
    select {}

}

