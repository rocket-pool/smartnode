package context

import (
	"math/big"
	"net/url"
	"os"

	"github.com/urfave/cli/v2"
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

// Add the Smart Node settings into a CLI context
func SetSmartNodeSettings(c *cli.Context, sn *SmartNodeSettings) {
	c.App.Metadata[contextMetadataName] = sn
}

// Get the Smart Node settings from a CLI context
func GetSmartNodeSettings(c *cli.Context) *SmartNodeSettings {
	return c.App.Metadata[contextMetadataName].(*SmartNodeSettings)
}
