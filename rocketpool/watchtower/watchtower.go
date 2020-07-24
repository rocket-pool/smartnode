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

    // Start tasks
    if err := startDissolveTimedOutMinipools(c); err != nil {
        return err
    }
    if err := startSubmitWithdrawableMinipools(c); err != nil {
        return err
    }
    if err := startSubmitNetworkBalances(c); err != nil {
        return err
    }
    if err := startProcessWithdrawals(c); err != nil {
        return err
    }

    // Block thread
    select {}

}

