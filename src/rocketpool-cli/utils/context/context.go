package context

import (
	"math/big"

	"github.com/urfave/cli/v2"
)

const (
	contextMetadataName string = "rp-context"
)

// Context for global settings
type SmartNodeContext struct {
	// The path to the configuration file
	ConfigPath string

	// The path to use for the daemon's API socket
	ApiSocketPath string

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
}

// Add the Smart Node context into a CLI context
func SetSmartnodeContext(c *cli.Context, sn *SmartNodeContext) {
	c.App.Metadata[contextMetadataName] = sn
}

// Get the Smart Node context from a CLI context
func GetSmartNodeContext(c *cli.Context) *SmartNodeContext {
	return c.App.Metadata[contextMetadataName].(*SmartNodeContext)
}
