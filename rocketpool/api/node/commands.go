package node

import (
    "regexp"
    "strconv"
    "strings"

    "github.com/ethereum/go-ethereum/common"
    "github.com/urfave/cli"

    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


// Register node subcommands
func RegisterSubcommands(command *cli.Command, name string, aliases []string) {
    command.Subcommands = append(command.Subcommands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Manage node registration & state",
        Subcommands: []cli.Command{

            // Get the node's status
            cli.Command{
                Name:      "status",
                Aliases:   []string{"s"},
                Usage:     "Get the node's status information",
                UsageText: "rocketpool node status",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    if err := cliutils.ValidateAPIArgs(c, 0, nil); err != nil {
                        return err
                    }

                    // Run command
                    return getNodeStatus(c)

                },
            },

            // Get the node's account address
            cli.Command{
                Name:      "account",
                Aliases:   []string{"c"},
                Usage:     "Get the node's account address",
                UsageText: "rocketpool node account",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    if err := cliutils.ValidateAPIArgs(c, 0, nil); err != nil {
                        return err
                    }

                    // Run command
                    return getNodeAccount(c)

                },
            },

            // Initialise the node password
            cli.Command{
                Name:      "canInitPassword",
                Usage:     "Can initialize the node password",
                UsageText: "rocketpool node canInitPassword",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    if err := cliutils.ValidateAPIArgs(c, 0, nil); err != nil {
                        return err
                    }

                    // Run command
                    return canInitNodePassword(c)

                },
            },
            cli.Command{
                Name:      "initPassword",
                Aliases:   []string{"p"},
                Usage:     "Initialize the node password",
                UsageText: "rocketpool node initPassword password",
                Action: func(c *cli.Context) error {

                    // Arguments
                    var password string

                    // Validate arguments
                    if err := cliutils.ValidateAPIArgs(c, 1, func(messages *[]string) {

                        // Check password
                        if password = c.Args().Get(0); len(password) < 8 {
                            *messages = append(*messages, "Password must be at least 8 characters long")
                        }

                    }); err != nil {
                        return err
                    }

                    // Run command
                    return initNodePassword(c, password)

                },
            },

            // Initialise the node account
            cli.Command{
                Name:      "canInitAccount",
                Usage:     "Can initialize the node account",
                UsageText: "rocketpool node canInitAccount",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    if err := cliutils.ValidateAPIArgs(c, 0, nil); err != nil {
                        return err
                    }

                    // Run command
                    return canInitNodeAccount(c)

                },
            },
            cli.Command{
                Name:      "initAccount",
                Aliases:   []string{"a"},
                Usage:     "Initialize the node account",
                UsageText: "rocketpool node initAccount",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    if err := cliutils.ValidateAPIArgs(c, 0, nil); err != nil {
                        return err
                    }

                    // Run command
                    return initNodeAccount(c)

                },
            },

            // Export the node account
            cli.Command{
                Name:      "export",
                Aliases:   []string{"e"},
                Usage:     "Export the node account",
                UsageText: "rocketpool node export",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    if err := cliutils.ValidateAPIArgs(c, 0, nil); err != nil {
                        return err
                    }

                    // Run command
                    return exportNodeAccount(c)

                },
            },

            // Register the node with Rocket Pool
            cli.Command{
                Name:      "canRegister",
                Usage:     "Can register the node on the Rocket Pool network",
                UsageText: "rocketpool node canRegister",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    if err := cliutils.ValidateAPIArgs(c, 0, nil); err != nil {
                        return err
                    }

                    // Run command
                    return canRegisterNode(c)

                },
            },
            cli.Command{
                Name:      "register",
                Aliases:   []string{"r"},
                Usage:     "Register the node on the Rocket Pool network",
                UsageText: "rocketpool node register timezone",
                Action: func(c *cli.Context) error {

                    // Arguments
                    var timezone string

                    // Validate arguments
                    if err := cliutils.ValidateAPIArgs(c, 1, func(messages *[]string) {

                        // Check timezone
                        if timezone = c.Args().Get(0); !regexp.MustCompile("^\\w{2,}\\/\\w{2,}$").MatchString(timezone) {
                            *messages = append(*messages, "Invalid timezone - must be in the format 'Country/City'")
                        }

                    }); err != nil {
                        return err
                    }

                    // Run command
                    return registerNode(c, timezone)

                },
            },

            // Withdraw resources from the node contract
            cli.Command{
                Name:      "withdraw",
                Aliases:   []string{"w"},
                Usage:     "Withdraw resources from the node's network contract",
                UsageText: "rocketpool node withdraw amount unit",
                Action: func(c *cli.Context) error {

                    // Arguments
                    var amount float64
                    var unit string

                    // Validate arguments
                    if err := cliutils.ValidateAPIArgs(c, 2, func(messages *[]string) {
                        var err error

                        // Parse amount
                        if amount, err = strconv.ParseFloat(c.Args().Get(0), 64); err != nil || amount <= 0 {
                            *messages = append(*messages, "Invalid amount - must be a positive decimal number")
                        }

                        // Parse unit
                        unit = strings.ToUpper(c.Args().Get(1))
                        switch unit {
                            case "ETH":
                            case "RPL":
                            default:
                                *messages = append(*messages, "Invalid unit - valid units are 'ETH' and 'RPL'")
                        }

                    }); err != nil {
                        return err
                    }

                    // Run command
                    return withdrawFromNode(c, amount, unit)

                },
            },

            // Send resources from the node account to an address
            cli.Command{
                Name:      "send",
                Aliases:   []string{"n"},
                Usage:     "Send resources from the node account to an address",
                UsageText: "rocketpool node send address amount unit",
                Action: func(c *cli.Context) error {

                    // Arguments
                    var address string
                    var amount float64
                    var unit string

                    // Validate arguments
                    if err := cliutils.ValidateAPIArgs(c, 3, func(messages *[]string) {
                        var err error

                        // Validate address
                        address = c.Args().Get(0)
                        if !common.IsHexAddress(address) {
                            *messages = append(*messages, "Invalid address - must be a valid Ethereum address")
                        }

                        // Parse amount
                        if amount, err = strconv.ParseFloat(c.Args().Get(1), 64); err != nil || amount <= 0 {
                            *messages = append(*messages, "Invalid amount - must be a positive decimal number")
                        }

                        // Parse unit
                        unit = strings.ToUpper(c.Args().Get(2))
                        switch unit {
                            case "ETH":
                            case "RETH":
                            case "RPL":
                            default:
                                *messages = append(*messages, "Invalid unit - valid units are 'ETH', 'RETH' and 'RPL'")
                        }

                    }); err != nil {
                        return err
                    }

                    // Run command
                    return sendFromNode(c, address, amount, unit)

                },
            },

            // Set the node's timezone
            cli.Command{
                Name:      "timezone",
                Aliases:   []string{"t"},
                Usage:     "Set the node's timezone on the Rocket Pool network",
                UsageText: "rocketpool node timezone tz",
                Action: func(c *cli.Context) error {

                    // Arguments
                    var timezone string

                    // Validate arguments
                    if err := cliutils.ValidateAPIArgs(c, 1, func(messages *[]string) {

                        // Check timezone
                        if timezone = c.Args().Get(0); !regexp.MustCompile("^\\w{2,}\\/\\w{2,}$").MatchString(timezone) {
                            *messages = append(*messages, "Invalid timezone - must be in the format 'Country/City'")
                        }

                    }); err != nil {
                        return err
                    }

                    // Run command
                    return setNodeTimezone(c, timezone)

                },
            },

        },
    })
}

