package storage

import (
    "encoding/hex"

    "github.com/urfave/cli"

    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
    hexutil "github.com/rocket-pool/smartnode/shared/utils/hex"
)


// Register storage subcommands
func RegisterSubcommands(command *cli.Command, name string, aliases []string) {
    command.Subcommands = append(command.Subcommands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Query data from RocketStorage",
        UsageText: "rocketpool storage type key",
        Action: func(c *cli.Context) error {

            // Arguments
            var dataType string
            var key [32]byte

            // Validate arguments
            if err := cliutils.ValidateAPIArgs(c, 2, func(messages *[]string) {

                // Get data type
                dataType = c.Args().Get(0)
                switch dataType {
                    case "address":
                    case "bool":
                    case "bytes":
                    case "bytes32":
                    case "int":
                    case "string":
                    case "uint":
                    default:
                        *messages = append(*messages, "Invalid data type - valid types are 'address', 'bool', 'bytes', 'bytes32', 'int', 'string', 'uint'")
                }

                // Get key
                if keyBytes, err := hex.DecodeString(hexutil.RemovePrefix(c.Args().Get(1))); err != nil {
                    *messages = append(*messages, "Invalid data key")
                } else if len(keyBytes) != 32 {
                    *messages = append(*messages, "Invalid data key length")
                } else {
                    copy(key[:], keyBytes[:])
                }

            }); err != nil {
                return err
            }

            // Run command
            return getStorage(c, dataType, key)

        },
    })
}

