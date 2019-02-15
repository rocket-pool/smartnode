package service

import (
    "fmt"
    "os"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/daemons"
    "github.com/rocket-pool/smartnode-cli/rocketpool/daemons/smartnode"
    cliutils "github.com/rocket-pool/smartnode-cli/rocketpool/utils/cli"
)


// Register user commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Manage daemon services",
        Subcommands: []cli.Command{

            // Smartnode service commands
            cli.Command{
                Name:      "smartnode",
                Aliases:   []string{"s"},
                Usage:     "Manage smartnode daemon service",
                Subcommands: serviceCommands("smartnode", smartnode.Run),
            },

        },
    })
}


// Get service commands
func serviceCommands(name string, run func(*cli.Context) error) []cli.Command {
    return []cli.Command{

        // Install daemon service
        cli.Command{
            Name:      "install",
            Aliases:   []string{"i"},
            Usage:     "Install " + name + " daemon service (using systemd); must be run as root",
            UsageText: "rocketpool service " + name + " install",
            Action: func(c *cli.Context) error {

                // Validate arguments
                if err := cliutils.ValidateArgs(c, 0, nil); err != nil {
                    return err
                }

                // Check user ID
                if os.Geteuid() != 0 {
                    return cli.NewExitError("Command must be run as root - try 'sudo rocketpool service " + name + " install'", 1)
                }

                // Run command
                if err := daemons.Install(name); err != nil {
                    return cli.NewExitError("The " + name + " daemon service could not be installed: " + err.Error(), 1)
                }

                // Return
                fmt.Println("The " + name + " daemon service was successfully installed")
                return nil

            },
        },

        // Uninstall daemon service
        cli.Command{
            Name:      "uninstall",
            Aliases:   []string{"u"},
            Usage:     "Uninstall " + name + " daemon service (using systemd); must be run as root",
            UsageText: "rocketpool service " + name + " uninstall",
            Action: func(c *cli.Context) error {

                // Validate arguments
                if err := cliutils.ValidateArgs(c, 0, nil); err != nil {
                    return err
                }

                // Check user ID
                if os.Geteuid() != 0 {
                    return cli.NewExitError("Command must be run as root - try 'sudo rocketpool service " + name + " uninstall'", 1)
                }

                // Run command
                if err := daemons.Uninstall(name); err != nil {
                    return cli.NewExitError("The " + name + " daemon service could not be uninstalled: " + err.Error(), 1)
                }

                // Return
                fmt.Println("The " + name + " daemon service was successfully uninstalled")
                return nil

            },
        },

        // Enable daemon service
        cli.Command{
            Name:      "enable",
            Aliases:   []string{"e"},
            Usage:     "Enable " + name + " daemon service to start at boot (using systemd); must be run as root",
            UsageText: "rocketpool service " + name + " enable",
            Action: func(c *cli.Context) error {

                // Validate arguments
                if err := cliutils.ValidateArgs(c, 0, nil); err != nil {
                    return err
                }

                // Check user ID
                if os.Geteuid() != 0 {
                    return cli.NewExitError("Command must be run as root - try 'sudo rocketpool service " + name + " enable'", 1)
                }

                // Run command
                if err := daemons.Enable(name); err != nil {
                    return cli.NewExitError("The " + name + " daemon service could not be enabled: " + err.Error(), 1)
                }

                // Return
                fmt.Println("The " + name + " daemon service was successfully enabled to start at boot")
                return nil

            },
        },

        // Disable daemon service
        cli.Command{
            Name:      "disable",
            Aliases:   []string{"d"},
            Usage:     "Disable " + name + " daemon service from starting at boot (using systemd); must be run as root",
            UsageText: "rocketpool service " + name + " disable",
            Action: func(c *cli.Context) error {

                // Validate arguments
                if err := cliutils.ValidateArgs(c, 0, nil); err != nil {
                    return err
                }

                // Check user ID
                if os.Geteuid() != 0 {
                    return cli.NewExitError("Command must be run as root - try 'sudo rocketpool service " + name + " disable'", 1)
                }

                // Run command
                if err := daemons.Disable(name); err != nil {
                    return cli.NewExitError("The " + name + " daemon service could not be disabled: " + err.Error(), 1)
                }

                // Return
                fmt.Println("The " + name + " daemon service was successfully disabled from starting at boot")
                return nil

            },
        },

        // Start daemon service
        cli.Command{
            Name:      "start",
            Aliases:   []string{"s"},
            Usage:     "Start " + name + " daemon service (using systemd); must be run as root",
            UsageText: "rocketpool service " + name + " start",
            Action: func(c *cli.Context) error {

                // Validate arguments
                if err := cliutils.ValidateArgs(c, 0, nil); err != nil {
                    return err
                }

                // Check user ID
                if os.Geteuid() != 0 {
                    return cli.NewExitError("Command must be run as root - try 'sudo rocketpool service " + name + " start'", 1)
                }

                // Run command
                if err := daemons.Start(name); err != nil {
                    return cli.NewExitError("The " + name + " daemon service could not be started: " + err.Error(), 1)
                }

                // Return
                fmt.Println("The " + name + " daemon service was successfully started")
                return nil

            },
        },

        // Stop daemon service
        cli.Command{
            Name:      "stop",
            Usage:     "Stop " + name + " daemon service (using systemd); must be run as root",
            UsageText: "rocketpool service " + name + " stop",
            Action: func(c *cli.Context) error {

                // Validate arguments
                if err := cliutils.ValidateArgs(c, 0, nil); err != nil {
                    return err
                }

                // Check user ID
                if os.Geteuid() != 0 {
                    return cli.NewExitError("Command must be run as root - try 'sudo rocketpool service " + name + " stop'", 1)
                }

                // Run command
                if err := daemons.Stop(name); err != nil {
                    return cli.NewExitError("The " + name + " daemon service could not be stopped: " + err.Error(), 1)
                }

                // Return
                fmt.Println("The " + name + " daemon service was successfully stopped")
                return nil

            },
        },

        // Get daemon status
        cli.Command{
            Name:      "status",
            Usage:     "Get " + name + " daemon service status (using systemd)",
            UsageText: "rocketpool service " + name + " status",
            Action: func(c *cli.Context) error {

                // Validate arguments
                if err := cliutils.ValidateArgs(c, 0, nil); err != nil {
                    return err
                }

                // Run command
                status := daemons.Status(name)

                // Return
                fmt.Print(status)
                return nil

            },
        },

        // Run daemon
        cli.Command{
            Name:      "run",
            Aliases:   []string{"r"},
            Usage:     "Run " + name + " daemon; for manual / advanced use only",
            UsageText: "rocketpool service " + name + " run",
            Action: func(c *cli.Context) error {

                // Validate arguments
                if err := cliutils.ValidateArgs(c, 0, nil); err != nil {
                    return err
                }

                // Run command
                return run(c)

            },
        },

    }
}

