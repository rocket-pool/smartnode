package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/urfave/cli/v2"
)

const (
	version    string = "1.5.1"
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
		{
			Name:  "Jacob Shufro",
			Email: "jacob@shuf.ro",
		},
	}
	app.Copyright = "(c) 2023 Rocket Pool Pty Ltd"

	// Set application flags
	app.Flags = []cli.Flag{
		&cli.Int64Flag{
			Name:    "interval",
			Aliases: []string{"i"},
			Usage:   "The rewards interval to generate the artifacts for. A value of -1 indicates that you want to do a \"dry run\" of generating the tree for the current (active) interval, using the current latest finalized block as the interval end (unless -t is passed).",
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
		&cli.Uint64Flag{
			Name:    "target-epoch",
			Aliases: []string{"t"},
			Usage:   "If provided, this flag will be used to override the last epoch of an interval, current or past. If passed with -i, the epoch must be part of the provided interval.",
		},
		&cli.Uint64Flag{
			Name:    "ruleset",
			Aliases: []string{"r"},
			Usage:   "The ruleset to use during generation. If not included, treegen will use the default ruleset for the network based on the rewards interval at the chosen block. Default of 0 will use whatever the ruleset specified by the network based on which block is being targeted.",
		},
		&cli.BoolFlag{
			Name:    "network-info",
			Aliases: []string{"n"},
			Usage:   "If provided, this will simply print out info about the network being used, the current or targeted interval, and the current or targeted ruleset.",
			Value:   false,
		},
		&cli.BoolFlag{
			Name:    "approximate-only",
			Aliases: []string{"a"},
			Usage:   "Approximates the rETH stakers' share of the Smoothing Pool at the current or target block instead of generating the entire rewards tree.",
			Value:   false,
		},
		&cli.BoolFlag{
			Name:    "use-rolling-records",
			Aliases: []string{"rr"},
			Usage:   "Enable the rolling record capability of the Smartnode tree generator. Use this to store and load record caches instead of recalculating attestation performance each time you run treegen.",
			Value:   false,
		},
		&cli.StringFlag{
			Name:    "cpuprofile",
			Aliases: []string{"c"},
			Usage:   "Path to which to save a pprof cpu profile, e.g. ./treegen.pprof. If unset, profiling is disabled.",
		},
		&cli.StringFlag{
			Name:    "memprofile",
			Aliases: []string{"m"},
			Usage:   "Path to which to save a pprof heap profile, e.g. ./treegen.pprof. If unset, profiling is disabled.",
		},
	}

	app.Action = func(c *cli.Context) error {
		cpuprofile := c.String("cpuprofile")
		if cpuprofile != "" {
			f, err := os.Create(cpuprofile)
			if err != nil {
				fmt.Printf("%sError generating tree: %s%s\n", colorRed, err.Error(), colorReset)
				os.Exit(1)
			}
			defer f.Close()
			if err := pprof.StartCPUProfile(f); err != nil {
				fmt.Printf("%sError generating tree: %s%s\n", colorRed, err.Error(), colorReset)
				os.Exit(1)
			}
			defer pprof.StopCPUProfile()
		}

		memprofile := c.String("memprofile")
		if memprofile != "" {
			defer func() {
				f, err := os.Create(memprofile)
				if err != nil {
					fmt.Printf("%sError saving heap profile: %w%w\n", colorRed, err, colorReset)
					os.Exit(1)
				}
				defer f.Close()
				runtime.GC()
				if err := pprof.WriteHeapProfile(f); err != nil {
					fmt.Printf("%sError saving heap profile: %w%w\n", colorRed, err, colorReset)
				}
			}()
		}

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
