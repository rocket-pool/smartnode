package main

import (
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/commands/auction"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/commands/faucet"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/commands/minipool"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/commands/network"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/commands/node"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/commands/odao"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/commands/pdao"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/commands/queue"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/commands/security"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/commands/service"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/commands/wallet"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/context"
	"github.com/rocket-pool/smartnode/v2/shared"
)

const (
	defaultConfigFolder string = ".rocketpool"
)

// Flags
var (
	allowRootFlag *cli.BoolFlag = &cli.BoolFlag{
		Name:    "allow-root",
		Aliases: []string{"r"},
		Usage:   "Allow rocketpool to be run as the root user",
	}
	configPathFlag *cli.StringFlag = &cli.StringFlag{
		Name:    "config-path",
		Aliases: []string{"c"},
		Usage:   "Directory to install and save all of Rocket Pool's configuration and data to",
	}
	nativeFlag *cli.BoolFlag = &cli.BoolFlag{
		Name:    "native-mode",
		Aliases: []string{"n"},
		Usage:   "Set this if you're running the Smart Node in Native Mode (where you manage your own Node process and don't use Docker for automatic management)",
	}
	maxFeeFlag *cli.Float64Flag = &cli.Float64Flag{
		Name:    "max-fee",
		Aliases: []string{"f"},
		Usage:   "The max fee (including the priority fee) you want a transaction to cost, in gwei. Use 0 to set it automatically based on network conditions.",
		Value:   0,
	}
	maxPriorityFeeFlag *cli.Float64Flag = &cli.Float64Flag{
		Name:    "max-priority-fee",
		Aliases: []string{"i"},
		Usage:   "The max priority fee you want a transaction to use, in gwei. Use 0 to set it automatically.",
		Value:   0,
	}
	nonceFlag *cli.Uint64Flag = &cli.Uint64Flag{
		Name:  "nonce",
		Usage: "Use this flag to explicitly specify the nonce that the next transaction should use, so it can override an existing 'stuck' transaction. If running a command that sends multiple transactions, the first will be given this nonce and the rest will be incremented sequentially.",
		Value: 0,
	}
	debugFlag *cli.BoolFlag = &cli.BoolFlag{
		Name:  "debug",
		Usage: "Enable debug printing of API commands",
	}
	secureSessionFlag *cli.BoolFlag = &cli.BoolFlag{
		Name:    "secure-session",
		Aliases: []string{"s"},
		Usage:   "Some commands may print sensitive information to your terminal. Use this flag when nobody can see your screen to allow sensitive data to be printed without prompting",
	}
)

// Run
func main() {
	// Add logo to application help template
	cli.AppHelpTemplate = fmt.Sprintf(`
______           _        _    ______           _ 
| ___ \         | |      | |   | ___ \         | |
| |_/ /___   ___| | _____| |_  | |_/ /__   ___ | |
|    // _ \ / __| |/ / _ \ __| |  __/ _ \ / _ \| |
| |\ \ (_) | (__|   <  __/ |_  | | | (_) | (_) | |
\_| \_\___/ \___|_|\_\___|\__| \_|  \___/ \___/|_|

%s`, cli.AppHelpTemplate)

	// Initialise application
	app := cli.NewApp()

	// Set application info
	app.Name = "rocketpool"
	app.Usage = "Smart Node CLI for Rocket Pool"
	app.Version = shared.RocketPoolVersion
	app.Authors = []*cli.Author{
		{
			Name:  "David Rugendyke",
			Email: "david@rocketpool.net",
		},
		{
			Name:  "Jake Pospischil",
			Email: "jake@rocketpool.net",
		},
		{
			Name:  "Joe Clapis",
			Email: "joe@rocketpool.net",
		},
		{
			Name:  "Kane Wallmann",
			Email: "kane@rocketpool.net",
		},
	}
	app.Copyright = "(c) 2024 Rocket Pool Pty Ltd"

	// Initialize app metadata
	app.Metadata = make(map[string]interface{})

	// Set application flags
	app.Flags = []cli.Flag{
		allowRootFlag,
		configPathFlag,
		nativeFlag,
		maxFeeFlag,
		maxPriorityFeeFlag,
		nonceFlag,
		utils.PrintTxDataFlag,
		utils.SignTxOnlyFlag,
		debugFlag,
		secureSessionFlag,
	}

	// Set default paths for flags before parsing the provided values
	setDefaultPaths()

	// Register commands
	auction.RegisterCommands(app, "auction", []string{"a"})
	faucet.RegisterCommands(app, "faucet", []string{"f"})
	minipool.RegisterCommands(app, "minipool", []string{"m"})
	network.RegisterCommands(app, "network", []string{"e"})
	node.RegisterCommands(app, "node", []string{"n"})
	odao.RegisterCommands(app, "odao", []string{"o"})
	pdao.RegisterCommands(app, "pdao", []string{"p"})
	queue.RegisterCommands(app, "queue", []string{"q"})
	security.RegisterCommands(app, "security", []string{"c"})
	service.RegisterCommands(app, "service", []string{"s"})
	wallet.RegisterCommands(app, "wallet", []string{"w"})

	app.Before = func(c *cli.Context) error {
		// Check user ID
		if os.Getuid() == 0 && !c.Bool(allowRootFlag.Name) {
			fmt.Fprintln(os.Stderr, "rocketpool should not be run as root. Please try again without 'sudo'.")
			fmt.Fprintf(os.Stderr, "If you want to run rocketpool as root anyway, use the '--%s' option to override this warning.\n", allowRootFlag.Name)
			os.Exit(1)
		}

		err := validateFlags(c)
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			os.Exit(1)
		}
		return nil
	}

	// Run application
	fmt.Println()
	if err := app.Run(os.Args); err != nil {
		utils.PrettyPrintError(err)
	}
	fmt.Println()
}

// Set the default paths for various flags
func setDefaultPaths() {
	// Get the home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Cannot get user's home directory: %s\n", err.Error())
		os.Exit(1)
	}

	// Default config folder path
	defaultConfigPath := filepath.Join(homeDir, defaultConfigFolder)
	configPathFlag.Value = defaultConfigPath
}

// Validate the global flags
func validateFlags(c *cli.Context) error {
	snCtx := &context.SmartNodeContext{
		MaxFee:         c.Float64(maxFeeFlag.Name),
		MaxPriorityFee: c.Float64(maxPriorityFeeFlag.Name),
		DebugEnabled:   c.Bool(debugFlag.Name),
		SecureSession:  c.Bool(secureSessionFlag.Name),
	}

	// If set, validate custom nonce
	snCtx.Nonce = big.NewInt(0)
	if c.IsSet(nonceFlag.Name) {
		customNonce := c.Uint64(nonceFlag.Name)
		snCtx.Nonce.SetUint64(customNonce)
	}

	// Make sure the config directory exists
	configPath := c.String(configPathFlag.Name)
	path, err := homedir.Expand(strings.TrimSpace(configPath))
	if err != nil {
		return fmt.Errorf("error expanding config path [%s]: %w", configPath, err)
	}
	snCtx.ConfigPath = path

	// Grab the daemon socket path; don't error out if it doesn't exist yet because this might be a new installation that hasn't configured and started it yet
	nativeMode := c.Bool(nativeFlag.Name)
	snCtx.NativeMode = nativeMode

	// TODO: more here
	context.SetSmartnodeContext(c, snCtx)
	return nil
}
