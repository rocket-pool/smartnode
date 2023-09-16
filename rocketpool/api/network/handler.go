package network

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/urfave/cli"

	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/rocketpool/api/handlers"
	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/rocketpool/common/services"
	"github.com/rocket-pool/smartnode/rocketpool/common/wallet"
	"github.com/rocket-pool/smartnode/shared/config"
	"github.com/rocket-pool/smartnode/shared/types/api"
	wtypes "github.com/rocket-pool/smartnode/shared/types/wallet"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// ===============
// === Handler ===
// ===============

// Context with services and common bindings for calls
type commonContext struct {
	w           *wallet.LocalWallet
	rp          *rocketpool.RocketPool
	cfg         *config.RocketPoolConfig
	opts        *bind.TransactOpts
	nodeAddress common.Address
}

// Register routes
func RegisterRoutes(router *mux.Router, name string) {
	server.RegisterSingleStageRoute(router, "dao-proposals", NewNetworkProposalHandler, runNetworkCall[api.NetworkDAOProposalsResponse])
	//server.RegisterSingleStageHandler(router, "withdraw-rpl", NewFaucetWithdrawHandler, runFaucetCall[api.FaucetWithdrawRplData])
}

// Register subcommands
func RegisterSubcommands(command *cli.Command, name string, aliases []string) {
	command.Subcommands = append(command.Subcommands, cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Manage Rocket Pool network parameters",
		Subcommands: []cli.Command{

			// Deposit contract info
			{
				Name:      "deposit-contract-info",
				Usage:     "Get information about the deposit contract specified by Rocket Pool and the Beacon Chain client",
				UsageText: "rocketpool api network deposit-contract-info",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(getDepositContractInfo(c))
					return nil

				},
			},

			{
				Name:      "node-fee",
				Aliases:   []string{"f"},
				Usage:     "Get the current network node commission rate",
				UsageText: "rocketpool api network node-fee",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(getNodeFee(c))
					return nil

				},
			},

			{
				Name:      "rpl-price",
				Aliases:   []string{"p"},
				Usage:     "Get the current network RPL price in ETH",
				UsageText: "rocketpool api network rpl-price",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(getRplPrice(c))
					return nil

				},
			},

			{
				Name:      "stats",
				Aliases:   []string{"s"},
				Usage:     "Get stats about the Rocket Pool network and its tokens",
				UsageText: "rocketpool api network stats",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(getStats(c))
					return nil

				},
			},

			{
				Name:      "timezone-map",
				Aliases:   []string{"t"},
				Usage:     "Get the table of node operators by timezone",
				UsageText: "rocketpool api network stats",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(getTimezones(c))
					return nil

				},
			},

			{
				Name:      "can-generate-rewards-tree",
				Usage:     "Check if the rewards tree for the provided interval can be generated",
				UsageText: "rocketpool api network can-generate-rewards-tree index",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}

					index, err := cliutils.ValidateUint("index", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canGenerateRewardsTree(c, index))
					return nil

				},
			},

			{
				Name:      "generate-rewards-tree",
				Usage:     "Set a request marker for the watchtower to generate the rewards tree for the given interval",
				UsageText: "rocketpool api network generate-rewards-tree index",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}

					index, err := cliutils.ValidateUint("index", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(generateRewardsTree(c, index))
					return nil

				},
			},

			{
				Name:      "dao-proposals",
				Aliases:   []string{"d"},
				Usage:     "Get the currently active DAO proposals",
				UsageText: "rocketpool api network dao-proposals",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(getActiveDAOProposals(c))
					return nil

				},
			},

			{
				Name:      "download-rewards-file",
				Aliases:   []string{"drf"},
				Usage:     "Download a rewards info file from IPFS for the given interval",
				UsageText: "rocketpool api service download-rewards-file interval",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}

					interval, err := cliutils.ValidatePositiveUint("interval", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(downloadRewardsFile(c, interval))
					return nil

				},
			},

			{
				Name:      "is-atlas-deployed",
				Aliases:   []string{"iad"},
				Usage:     "Checks if Atlas has been deployed yet.",
				UsageText: "rocketpool api network is-atlas-deployed",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(isAtlasDeployed(c))
					return nil

				},
			},

			{
				Name:      "latest-delegate",
				Usage:     "Get the address of the latest minipool delegate contract.",
				UsageText: "rocketpool api network latest-delegate",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(getLatestDelegate(c))
					return nil

				},
			},
		},
	})
}

// Create a scaffolded generic call handler, with caller-specific functionality where applicable
func runNetworkCall[dataType any](h handlers.ISingleStageCallContext[dataType, callContext]) (*api.ApiResponse[dataType], error) {
	// Get services
	if err := services.RequireNodeRegistered(); err != nil {
		return nil, fmt.Errorf("error checking if node is registered: %w", err)
	}
	sp := services.GetServiceProvider()
	w := sp.GetWallet()
	rp := sp.GetRocketPool()
	cfg := sp.GetConfig()
	address, _ := w.GetAddress()

	// Get the transact opts if this node is ready for transaction
	var opts *bind.TransactOpts
	walletStatus := w.GetStatus()
	if walletStatus == wtypes.WalletStatus_Ready {
		var err error
		opts, err = w.GetTransactor()
		if err != nil {
			return nil, fmt.Errorf("error getting node account transactor: %w", err)
		}
	}

	// Response
	data := new(dataType)
	response := &api.ApiResponse[dataType]{
		WalletStatus: walletStatus,
		Data:         data,
	}

	// Create the context
	context := &commonContext{
		w:           w,
		rp:          rp,
		cfg:         cfg,
		opts:        opts,
		nodeAddress: address,
	}

	// Supplemental function-specific bindings
	err := h.CreateBindings(context)
	if err != nil {
		return nil, err
	}

	// Get contract state
	err = rp.Query(func(mc *batch.MultiCaller) error {
		h.GetState(context, mc)
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting contract state: %w", err)
	}

	// Supplemental function-specific response construction
	err = h.PrepareData(context, data)
	if err != nil {
		return nil, err
	}

	// Return
	return response, nil
}
