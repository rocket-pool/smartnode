package security

import "github.com/urfave/cli/v2"

var proposalFlag *cli.Uint64Flag = &cli.Uint64Flag{
	Name:    "proposal",
	Aliases: []string{"p"},
}
