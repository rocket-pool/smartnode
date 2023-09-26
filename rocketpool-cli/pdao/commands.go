package pdao

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

const (
	boolUsage string = "accepts 'true', 'false', 'yes', or 'no'"
)

// Register commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
	app.Commands = append(app.Commands, cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Manage the Rocket Pool Protocol DAO",
		Subcommands: []cli.Command{

			{
				Name:      "settings",
				Aliases:   []string{"s"},
				Usage:     "Show all of the current Protocol DAO settings and values",
				UsageText: "rocketpool pdao settings",
				Action: func(c *cli.Context) error {

					// Run
					return getSettings(c)

				},
			},

			{
				Name:    "propose",
				Aliases: []string{"p"},
				Usage:   "Make a Protocol DAO proposal",
				Subcommands: []cli.Command{

					{
						Name:    "setting",
						Aliases: []string{"s"},
						Usage:   "Make a Protocol DAO setting proposal",
						Subcommands: []cli.Command{

							{
								Name:    "auction",
								Aliases: []string{"a"},
								Usage:   "Auction settings",
								Subcommands: []cli.Command{

									{
										Name:      "is-create-lot-enabled",
										Aliases:   []string{"icle"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.CreateLotEnabledSettingPath, boolUsage),
										UsageText: "rocketpool pdao propose setting auction is-create-lot-enabled value",
										Action: func(c *cli.Context) error {

											// Validate args
											if err := cliutils.ValidateArgCount(c, 1); err != nil {
												return err
											}
											value, err := cliutils.ValidateBool("value", c.Args().Get(0))
											if err != nil {
												return err
											}

											// Run
											return proposeSettingAuctionIsCreateLotEnabled(c, value)

										},
									},
								},
							},
						},
					},
				},
			},

			{
				Name:    "proposals",
				Aliases: []string{"o"},
				Usage:   "Manage oracle DAO proposals",
				Subcommands: []cli.Command{

					{
						Name:      "list",
						Aliases:   []string{"l"},
						Usage:     "List the oracle DAO proposals",
						UsageText: "rocketpool pdao proposals list",
						Flags: []cli.Flag{
							cli.StringFlag{
								Name:  "states, s",
								Usage: "Comma separated list of states to filter ('pending', 'active', 'succeeded', 'executed', 'cancelled', 'defeated', or 'expired')",
								Value: "",
							},
						},
						Action: func(c *cli.Context) error {

							// Validate args
							if err := cliutils.ValidateArgCount(c, 0); err != nil {
								return err
							}

							// Run
							return getProposals(c, c.String("states"))

						},
					},

					{
						Name:      "details",
						Aliases:   []string{"d"},
						Usage:     "View proposal details",
						UsageText: "rocketpool pdao proposals details proposal-id",
						Action: func(c *cli.Context) error {

							// Validate args
							var err error
							if err = cliutils.ValidateArgCount(c, 1); err != nil {
								return err
							}
							id, err := cliutils.ValidateUint("proposal-id", c.Args().Get(0))
							if err != nil {
								return err
							}

							// Run
							return getProposal(c, id)

						},
					},

					{
						Name:      "cancel",
						Aliases:   []string{"c"},
						Usage:     "Cancel a proposal made by the node",
						UsageText: "rocketpool pdao proposals cancel [options]",
						Flags: []cli.Flag{
							cli.StringFlag{
								Name:  "proposal, p",
								Usage: "The ID of the proposal to cancel",
							},
						},
						Action: func(c *cli.Context) error {

							// Validate args
							if err := cliutils.ValidateArgCount(c, 0); err != nil {
								return err
							}

							// Validate flags
							if c.String("proposal") != "" {
								if _, err := cliutils.ValidatePositiveUint("proposal ID", c.String("proposal")); err != nil {
									return err
								}
							}

							// Run
							return cancelProposal(c)

						},
					},

					{
						Name:      "vote",
						Aliases:   []string{"v"},
						Usage:     "Vote on a proposal",
						UsageText: "rocketpool pdao proposals vote [options]",
						Flags: []cli.Flag{
							cli.StringFlag{
								Name:  "proposal, p",
								Usage: "The ID of the proposal to vote on",
							},
							cli.StringFlag{
								Name:  "support, s",
								Usage: "Whether to support the proposal ('yes' or 'no')",
							},
							cli.BoolFlag{
								Name:  "yes, y",
								Usage: "Automatically confirm vote",
							},
						},
						Action: func(c *cli.Context) error {

							// Validate args
							if err := cliutils.ValidateArgCount(c, 0); err != nil {
								return err
							}

							// Validate flags
							if c.String("proposal") != "" {
								if _, err := cliutils.ValidatePositiveUint("proposal ID", c.String("proposal")); err != nil {
									return err
								}
							}
							if c.String("support") != "" {
								if _, err := cliutils.ValidateBool("support", c.String("support")); err != nil {
									return err
								}
							}

							// Run
							return voteOnProposal(c)

						},
					},

					{
						Name:      "execute",
						Aliases:   []string{"x"},
						Usage:     "Execute a proposal",
						UsageText: "rocketpool pdao proposals execute [options]",
						Flags: []cli.Flag{
							cli.StringFlag{
								Name:  "proposal, p",
								Usage: "The ID of the proposal to execute (or 'all')",
							},
						},
						Action: func(c *cli.Context) error {

							// Validate args
							if err := cliutils.ValidateArgCount(c, 0); err != nil {
								return err
							}

							// Validate flags
							if c.String("proposal") != "" && c.String("proposal") != "all" {
								if _, err := cliutils.ValidatePositiveUint("proposal ID", c.String("proposal")); err != nil {
									return err
								}
							}

							// Run
							return executeProposal(c)

						},
					},
				},
			},
		},
	})
}
