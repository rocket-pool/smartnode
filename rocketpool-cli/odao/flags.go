package odao

import "github.com/urfave/cli/v2"

var memberFlag *cli.StringFlag = &cli.StringFlag{
	Name:    "member",
	Aliases: []string{"m"},
}

var proposalFlag *cli.Uint64Flag = &cli.Uint64Flag{
	Name:    "proposal",
	Aliases: []string{"p"},
}
