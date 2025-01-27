package debug

import (
	"fmt"

	"github.com/urfave/cli"

	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// Register subcommands
func RegisterSubcommands(command *cli.Command, name string, aliases []string) {
	command.Subcommands = append(command.Subcommands, cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Debugging and troubleshooting commands",
		Subcommands: []cli.Command{

			{
				Name:      "export-validators",
				Aliases:   []string{"x"},
				Usage:     "Exports a TSV file of validators",
				UsageText: "rocketpool api debug export-validators",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Export TSV of validators
					if err := ExportValidators(c); err != nil {
						fmt.Printf("An error occurred: %s\n", err)
					}
					return nil

				},
			},
			{
				Name:      "get-beacon-state",
				Aliases:   []string{"b"},
				Usage:     "Returns the beacon state for a given slot number",
				UsageText: "rocketpool api debug get-beacon-state slot-number",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}

					slotNumber, err := cliutils.ValidatePositiveUint("slot number", c.Args().Get(0))
					if err != nil {
						return err
					}

					validatorIndex, err := cliutils.ValidatePositiveUint("validator index", c.Args().Get(1))
					if err != nil {
						return err
					}

					if err := getBeaconStateForSlot(c, slotNumber, validatorIndex); err != nil {
						fmt.Printf("An error occurred: %s\n", err)
					}
					return nil

				},
			},
		},
	})
}
