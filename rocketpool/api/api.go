package api

import (
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/rocketpool/api/debug"
	"github.com/rocket-pool/smartnode/rocketpool/api/pdao"
	"github.com/rocket-pool/smartnode/rocketpool/api/security"
	"github.com/urfave/cli"

	"github.com/rocket-pool/rocketpool-go/utils"
	"github.com/rocket-pool/smartnode/rocketpool/api/auction"
	"github.com/rocket-pool/smartnode/rocketpool/api/minipool"
	"github.com/rocket-pool/smartnode/rocketpool/api/network"
	"github.com/rocket-pool/smartnode/rocketpool/api/node"
	"github.com/rocket-pool/smartnode/rocketpool/api/odao"
	"github.com/rocket-pool/smartnode/rocketpool/api/queue"
	apiservice "github.com/rocket-pool/smartnode/rocketpool/api/service"
	"github.com/rocket-pool/smartnode/rocketpool/api/wallet"
	"github.com/rocket-pool/smartnode/shared/services"
	apitypes "github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

const (
	MaxConcurrentEth1Requests = 200
)

// Waits for an auction transaction
func waitForTransaction(c *cli.Context, hash common.Hash) (*apitypes.APIResponse, error) {

	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := apitypes.APIResponse{}
	_, err = utils.WaitForTransaction(rp.Client, hash)
	if err != nil {
		return nil, err
	}

	// Return response
	return &response, nil

}

// Register commands
func RegisterCommands(app *cli.App, name string, aliases []string) {

	// CLI command
	command := cli.Command{
		Name:        name,
		Aliases:     aliases,
		Usage:       "Run Rocket Pool API commands",
		Subcommands: []cli.Command{},
	}

	// Don't show help message for api errors because of JSON serialisation
	command.OnUsageError = func(context *cli.Context, err error, isSubcommand bool) error {
		return err
	}

	// Register subcommands
	auction.RegisterSubcommands(&command, "auction", []string{"a"})
	minipool.RegisterSubcommands(&command, "minipool", []string{"m"})
	network.RegisterSubcommands(&command, "network", []string{"e"})
	node.RegisterSubcommands(&command, "node", []string{"n"})
	odao.RegisterSubcommands(&command, "odao", []string{"o"})
	pdao.RegisterSubcommands(&command, "pdao", []string{"p"})
	queue.RegisterSubcommands(&command, "queue", []string{"q"})
	security.RegisterSubcommands(&command, "security", []string{"c"})
	apiservice.RegisterSubcommands(&command, "service", []string{"s"})
	wallet.RegisterSubcommands(&command, "wallet", []string{"w"})
	debug.RegisterSubcommands(&command, "debug", []string{"d"})

	// Append a general wait-for-transaction command to support async operations
	command.Subcommands = append(command.Subcommands, cli.Command{
		Name:      "wait",
		Aliases:   []string{"t"},
		Usage:     "Wait for a transaction to complete",
		UsageText: "rocketpool api wait tx-hash",
		Action: func(c *cli.Context) error {
			// Validate args
			if err := cliutils.ValidateArgCount(c, 1); err != nil {
				return err
			}
			hash, err := cliutils.ValidateTxHash("tx-hash", c.Args().Get(0))
			if err != nil {
				return err
			}

			// Run
			api.PrintResponse(waitForTransaction(c, hash))
			return nil
		},
	})

	// Register CLI command
	app.Commands = append(app.Commands, command)

	// The daemon makes a large number of concurrent RPC requests to the Eth1 client
	// The HTTP transport is set to cache connections for future re-use equal to the maximum expected number of concurrent requests
	// This prevents issues related to memory consumption and address allowance from repeatedly opening and closing connections
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = MaxConcurrentEth1Requests

}
