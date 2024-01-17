package pdao

import "github.com/urfave/cli/v2"

var proposalFlag *cli.Uint64Flag = &cli.Uint64Flag{
	Name:    "proposal",
	Aliases: []string{"p"},
}

var recipientFlag *cli.StringFlag = &cli.StringFlag{
	Name:    "recipient",
	Aliases: []string{"r"},
	Usage:   "The recipient of the spend",
}

var amountFlag *cli.StringFlag = &cli.StringFlag{
	Name:    "amount",
	Aliases: []string{"a"},
	Usage:   "The amount of RPL to send",
}

var contractNameFlag *cli.StringFlag = &cli.StringFlag{
	Name:    "contract-name",
	Aliases: []string{"c"},
	Usage:   "The name of the recurring spend's contract / invoice (alternatively, the name of the recipient)",
}

var amountPerPeriodFlag *cli.StringFlag = &cli.StringFlag{
	Name:    "amount-per-period",
	Aliases: []string{"a"},
	Usage:   "The amount of RPL to send",
}

var periodLengthFlag *cli.StringFlag = &cli.StringFlag{
	Name:    "period-length",
	Aliases: []string{"l"},
	Usage:   "The length of time between each payment, in hours / minutes / seconds (e.g., 168h0m0s)",
}

var numberOfPeriodsFlag *cli.Uint64Flag = &cli.Uint64Flag{
	Name:    "number-of-periods",
	Aliases: []string{"p"},
	Usage:   "The total number of payment periods for the spend",
}

var voteDirectionFlag *cli.StringFlag = &cli.StringFlag{
	Name:    "vote-direction",
	Aliases: []string{"v"},
	Usage:   "How to vote ('abstain', 'for', 'against', 'veto')",
}
