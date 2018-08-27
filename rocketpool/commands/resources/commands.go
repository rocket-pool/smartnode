package resources

import (
    "fmt"
    "strconv"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/commands"
)


// Register resource commands
func RegisterCommands(app *cli.App, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      "resource",
        Aliases:   aliases,
        Usage:     "Manage resources",
        Subcommands: []cli.Command{

            // Check free resources
            cli.Command{
                Name:      "free",
                Aliases:   []string{"f"},
                Usage:     "Check free resources assigned to the node",
                UsageText: "rocketpool resource free [type]" + "\n   " +
                           "- valid types are 'eth' and 'rpl'",
                Action: func(c *cli.Context) error {

                    // Arguments
                    var resourceType string

                    // Validate arguments
                    err := commands.ValidateArgs(c, 1, func(messages *[]string) {

                        // Parse type
                        resourceType = c.Args().Get(0)
                        switch resourceType {
                            case "eth":
                            case "rpl":
                            default:
                                *messages = append(*messages, "Invalid type - valid types are 'eth' and 'rpl'")
                        }

                    });
                    if err != nil {
                        return err;
                    }

                    // Run command
                    fmt.Printf("Free %v: 0\n", resourceType)
                    return nil

                },
            },

            // Check used resources
            cli.Command{
                Name:      "used",
                Aliases:   []string{"u"},
                Usage:     "Check used resources assigned to the node",
                UsageText: "rocketpool resource used [type]" + "\n   " +
                           "- valid types are 'eth' and 'rpl'",
                Action: func(c *cli.Context) error {

                    // Arguments
                    var resourceType string

                    // Validate arguments
                    err := commands.ValidateArgs(c, 1, func(messages *[]string) {

                        // Parse type
                        resourceType = c.Args().Get(0)
                        switch resourceType {
                            case "eth":
                            case "rpl":
                            default:
                                *messages = append(*messages, "Invalid type - valid types are 'eth' and 'rpl'")
                        }

                    });
                    if err != nil {
                        return err;
                    }

                    // Run command
                    fmt.Printf("Used %v: 0\n", resourceType)
                    return nil

                },
            },

            // Check resources required
            cli.Command{
                Name:      "required",
                Aliases:   []string{"r"},
                Usage:     "Check resources required based on current network utilisation",
                UsageText: "rocketpool resource required [type, ether amount]" + "\n   " +
                           "- valid types are 'rpl'" + "\n   " +
                           "- ether amount must be a decimal number",
                Action: func(c *cli.Context) error {

                    // Arguments
                    var resourceType string
                    var etherAmount float64

                    // Validate arguments
                    err := commands.ValidateArgs(c, 2, func(messages *[]string) {
                        var err error

                        // Parse type
                        resourceType = c.Args().Get(0)
                        switch resourceType {
                            case "rpl":
                            default:
                                *messages = append(*messages, "Invalid type - valid types are 'rpl'")
                        }

                        // Parse ether amount
                        etherAmount, err = strconv.ParseFloat(c.Args().Get(1), 64)
                        if err != nil {
                            *messages = append(*messages, "Invalid ether amount - must be a decimal number")
                        }

                    });
                    if err != nil {
                        return err;
                    }

                    // Run command
                    fmt.Printf("Required %v for %v eth: 0\n", resourceType, etherAmount)
                    return nil

                },
            },

        },
    })
}

