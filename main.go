package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

const (
	version    string = "0.1.0"
	colorReset string = "\033[0m"
	colorRed   string = "\033[31m"
)

func main() {

	// Initialise application
	app := cli.NewApp()

	// Set application info
	app.Name = "treegen"
	app.Usage = "This application can be used to generate past Merkle rewards trees for the Rocket Pool network, or preview / test generation of the tree for the current interval."
	app.Version = version
	app.Authors = []*cli.Author{
		{
			Name:  "Joe Clapis",
			Email: "joe@rocketpool.net",
		},
	}
	app.Copyright = "(c) 2022 Rocket Pool Pty Ltd"

	// Set application flags
	app.Flags = []cli.Flag{
		&cli.Int64Flag{
			Name:    "interval",
			Aliases: []string{"i"},
			Usage:   "The rewards interval to generate the artifacts for. A value of -1 indicates that you want to do a \"dry run\" of generating the tree for the current (active) interval, using the current latest finalized block as the interval end.",
			Value:   -1,
		},
		&cli.StringFlag{
			Name:    "ec-endpoint",
			Aliases: []string{"e"},
			Usage:   "The URL of the Execution Client's JSON-RPC API. Note that for past interval generation, this must be an Archive EC.",
			Value:   "http://localhost:8545",
		},
		&cli.StringFlag{
			Name:    "bn-endpoint",
			Aliases: []string{"b"},
			Usage:   "The URL of the Beacon Node's REST API. Note that for past interval generation, this must have Archive capability (ability to replay arbitrary historical states).",
			Value:   "http://localhost:5052",
		},
		&cli.StringFlag{
			Name:    "output-dir",
			Aliases: []string{"o"},
			Usage:   "Output directory to save generated files.",
		},
		&cli.BoolFlag{
			Name:    "pretty-print",
			Aliases: []string{"p"},
			Usage:   "Toggle for saving the files in pretty-print format so they're human readable.",
			Value:   true,
		},
	}

	app.Action = func(c *cli.Context) error {
		return GenerateTree(c)
	}

	// Run application
	fmt.Println("")
	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("%sError generating tree: %s%s\n", colorRed, err.Error(), colorReset)
		os.Exit(1)
	}
	fmt.Println("")

}
