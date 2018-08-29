package service

import (
    "fmt"
    "os"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/commands"
    "github.com/rocket-pool/smartnode-cli/rocketpool/daemon"
)


// Register user commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Manage smartnode daemon service",
        Subcommands: []cli.Command{

            // Install daemon service
            cli.Command{
                Name:      "install",
                Aliases:   []string{"i"},
                Usage:     "Install smartnode daemon service (using systemd); must be run as root",
                UsageText: "rocketpool service install",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    err := commands.ValidateArgs(c, 0, nil)
                    if err != nil {
                        return err
                    }

                    // Check user ID
                    id := os.Geteuid()
                    if id != 0 {
                        return cli.NewExitError("Command must be run as root - try 'sudo rocketpool service install'", 1)
                    }

                    // Run command
                    err = daemon.Install()
                    if err != nil {
                        return cli.NewExitError("The smartnode daemon service could not be installed: " + err.Error(), 1)
                    }

                    // Return
                    fmt.Println("The smartnode daemon service was successfully installed")
                    return nil

                },
            },

            // Uninstall daemon service
            cli.Command{
                Name:      "uninstall",
                Aliases:   []string{"u"},
                Usage:     "Uninstall smartnode daemon service (using systemd); must be run as root",
                UsageText: "rocketpool service uninstall",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    err := commands.ValidateArgs(c, 0, nil)
                    if err != nil {
                        return err
                    }

                    // Check user ID
                    id := os.Geteuid()
                    if id != 0 {
                        return cli.NewExitError("Command must be run as root - try 'sudo rocketpool service uninstall'", 1)
                    }

                    // Run command
                    err = daemon.Uninstall()
                    if err != nil {
                        return cli.NewExitError("The smartnode daemon service could not be uninstalled: " + err.Error(), 1)
                    }

                    // Return
                    fmt.Println("The smartnode daemon service was successfully uninstalled")
                    return nil

                },
            },

            // Enable daemon service
            cli.Command{
                Name:      "enable",
                Aliases:   []string{"e"},
                Usage:     "Enable smartnode daemon service to start at boot (using systemd); must be run as root",
                UsageText: "rocketpool service enable",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    err := commands.ValidateArgs(c, 0, nil)
                    if err != nil {
                        return err
                    }

                    // Check user ID
                    id := os.Geteuid()
                    if id != 0 {
                        return cli.NewExitError("Command must be run as root - try 'sudo rocketpool service enable'", 1)
                    }

                    // Run command
                    err = daemon.Enable()
                    if err != nil {
                        return cli.NewExitError("The smartnode daemon service could not be enabled: " + err.Error(), 1)
                    }

                    // Return
                    fmt.Println("The smartnode daemon service was successfully enabled to start at boot")
                    return nil

                },
            },

            // Disable daemon service
            cli.Command{
                Name:      "disable",
                Aliases:   []string{"d"},
                Usage:     "Disable smartnode daemon service from starting at boot (using systemd); must be run as root",
                UsageText: "rocketpool service disable",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    err := commands.ValidateArgs(c, 0, nil)
                    if err != nil {
                        return err
                    }

                    // Check user ID
                    id := os.Geteuid()
                    if id != 0 {
                        return cli.NewExitError("Command must be run as root - try 'sudo rocketpool service disable'", 1)
                    }

                    // Run command
                    err = daemon.Disable()
                    if err != nil {
                        return cli.NewExitError("The smartnode daemon service could not be disabled: " + err.Error(), 1)
                    }

                    // Return
                    fmt.Println("The smartnode daemon service was successfully disabled from starting at boot")
                    return nil

                },
            },

            // Start daemon service
            cli.Command{
                Name:      "start",
                Aliases:   []string{"s"},
                Usage:     "Start smartnode daemon service (using systemd); must be run as root",
                UsageText: "rocketpool service start",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    err := commands.ValidateArgs(c, 0, nil)
                    if err != nil {
                        return err
                    }

                    // Check user ID
                    id := os.Geteuid()
                    if id != 0 {
                        return cli.NewExitError("Command must be run as root - try 'sudo rocketpool service start'", 1)
                    }

                    // Run command
                    err = daemon.Start()
                    if err != nil {
                        return cli.NewExitError("The smartnode daemon service could not be started: " + err.Error(), 1)
                    }

                    // Return
                    fmt.Println("The smartnode daemon service was successfully started")
                    return nil

                },
            },

            // Stop daemon service
            cli.Command{
                Name:      "stop",
                Usage:     "Stop smartnode daemon service (using systemd); must be run as root",
                UsageText: "rocketpool service stop",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    err := commands.ValidateArgs(c, 0, nil)
                    if err != nil {
                        return err
                    }

                    // Check user ID
                    id := os.Geteuid()
                    if id != 0 {
                        return cli.NewExitError("Command must be run as root - try 'sudo rocketpool service stop'", 1)
                    }

                    // Run command
                    err = daemon.Stop()
                    if err != nil {
                        return cli.NewExitError("The smartnode daemon service could not be stopped: " + err.Error(), 1)
                    }

                    // Return
                    fmt.Println("The smartnode daemon service was successfully stopped")
                    return nil

                },
            },

            // Get daemon status
            cli.Command{
                Name:      "status",
                Usage:     "Get smartnode daemon service status (using systemd)",
                UsageText: "rocketpool service status",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    err := commands.ValidateArgs(c, 0, nil)
                    if err != nil {
                        return err
                    }

                    // Run command
                    status := daemon.Status()

                    // Return
                    fmt.Println(status)
                    return nil

                },
            },

            // Run daemon
            cli.Command{
                Name:      "run",
                Aliases:   []string{"r"},
                Usage:     "Run smartnode daemon; for manual / advanced use only",
                UsageText: "rocketpool service run",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    err := commands.ValidateArgs(c, 0, nil)
                    if err != nil {
                        return err
                    }

                    // Run command
                    daemon.Run()
                    return nil

                },
            },

        },
    })
}

