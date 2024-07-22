package settings

import (
	"fmt"
	"math/big"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/urfave/cli/v2"
)

const (
	traceMode           os.FileMode = 0644
	defaultConfigFolder string      = ".rocketpool"
)

var (
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
	apiAddressFlag *cli.StringFlag = &cli.StringFlag{
		Name:    "api-address",
		Aliases: []string{"a"},
		Usage:   "The address of the Smart Node API server to connect to. If left blank it will default to 'localhost' at the port specified in the service configuration.",
	}
	httpTracePathFlag *cli.StringFlag = &cli.StringFlag{
		Name:    "http-trace-path",
		Aliases: []string{"htp"},
		Usage:   "The path to save HTTP trace logs to. Leave blank to disable HTTP tracing",
	}
)

const (
	contextMetadataName string = "rp-context"
)

// Context for global settings
type SmartNodeSettings struct {
	// The path to the configuration file
	ConfigPath string

	// True if this CLI should be run in Native Mode
	NativeMode bool

	// The max fee for transactions
	MaxFee float64

	// The max priority fee for transactions
	MaxPriorityFee float64

	// The nonce for the first transaction, if set
	Nonce *big.Int

	// True if debug mode is enabled
	DebugEnabled bool

	// True if this is a secure session
	SecureSession bool

	// The address and URL of the API server
	ApiUrl *url.URL

	// The HTTP trace file if tracing is enabled
	HttpTraceFile *os.File
}

// Get the Smart Node settings from a CLI context
func GetSmartNodeSettings(c *cli.Context) *SmartNodeSettings {
	return c.App.Metadata[contextMetadataName].(*SmartNodeSettings)
}

func AppendSmartNodeSettingsFlags(flags []cli.Flag) []cli.Flag {
	// Set default paths for flags before parsing the provided values
	setDefaultPaths()

	return append(flags,
		configPathFlag,
		nativeFlag,
		apiAddressFlag,
		maxFeeFlag,
		maxPriorityFeeFlag,
		nonceFlag,
		debugFlag,
		httpTracePathFlag,
		secureSessionFlag,
	)
}

// Validate the global flags
func NewSmartNodeSettings(c *cli.Context) (*SmartNodeSettings, error) {
	snSettings := &SmartNodeSettings{
		MaxFee:         c.Float64(maxFeeFlag.Name),
		MaxPriorityFee: c.Float64(maxPriorityFeeFlag.Name),
		DebugEnabled:   c.Bool(debugFlag.Name),
		SecureSession:  c.Bool(secureSessionFlag.Name),
	}

	// If set, validate custom nonce
	snSettings.Nonce = big.NewInt(0)
	if c.IsSet(nonceFlag.Name) {
		customNonce := c.Uint64(nonceFlag.Name)
		snSettings.Nonce.SetUint64(customNonce)
	}

	// Make sure the config directory exists
	configPath := c.String(configPathFlag.Name)
	path, err := homedir.Expand(strings.TrimSpace(configPath))
	if err != nil {
		return nil, fmt.Errorf("error expanding config path [%s]: %w", configPath, err)
	}
	snSettings.ConfigPath = path

	// Grab the daemon socket path; don't error out if it doesn't exist yet because this might be a new installation that hasn't configured and started it yet
	nativeMode := c.Bool(nativeFlag.Name)
	snSettings.NativeMode = nativeMode

	// Get the API URL
	address := c.String(apiAddressFlag.Name)
	if address != "" {
		baseUrl, err := url.Parse(address)
		if err != nil {
			return nil, fmt.Errorf("error parsing API address [%s]: %w", snSettings.ApiUrl, err)
		}
		snSettings.ApiUrl = baseUrl.JoinPath(config.SmartNodeApiClientRoute)
	}

	// Get the HTTP trace flag
	httpTracePath := c.String(httpTracePathFlag.Name)
	if httpTracePath != "" {
		snSettings.HttpTraceFile, err = os.OpenFile(httpTracePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, traceMode)
		if err != nil {
			return nil, fmt.Errorf("error opening HTTP trace file [%s]: %w", httpTracePath, err)
		}
	}

	c.App.Metadata[contextMetadataName] = snSettings

	return snSettings, nil
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
