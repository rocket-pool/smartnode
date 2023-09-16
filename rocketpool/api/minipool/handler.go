package minipool

import (
	"github.com/gorilla/mux"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/rocketpool/common/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

type MinipoolHandler struct {
	serviceProvider               *services.ServiceProvider
	beginReduceBondDetailsFactory server.IMinipoolCallContextFactory[*minipoolBeginReduceBondDetailsContext, api.MinipoolBeginReduceBondDetailsData]
}

func NewMinipoolHandler(serviceProvider *services.ServiceProvider) *MinipoolHandler {
	h := &MinipoolHandler{
		serviceProvider: serviceProvider,
	}
	h.beginReduceBondDetailsFactory = &minipoolBeginReduceBondDetailsContextFactory{h}
	return h
}

func (h *MinipoolHandler) RegisterRoutes(router *mux.Router) {
	server.RegisterMinipoolRoute(router, "begin-reduce-bond/details", h.beginReduceBondDetailsFactory, h.serviceProvider)
}

// Register subcommands
func RegisterSubcommands(command *cli.Command, name string, aliases []string) {
	command.Subcommands = append(command.Subcommands, cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Manage the node's minipools",
		Subcommands: []cli.Command{

			// Begin reduce bond
			{
				Name:      "get-minipool-begin-reduce-bond-details",
				Usage:     "Check whether any of the minipools belonging to the node can begin the bond reduction process",
				UsageText: "rocketpool api minipool get-minipool-begin-reduce-bond-details new-bond-amount-wei",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					newBondAmountWei, err := cliutils.ValidateWeiAmount("new bond amount", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					response, err := runMinipoolQuery[types.MinipoolBeginReduceBondDetailsData](c, &minipoolBeginReduceBondManager{
						newBondAmountWei: newBondAmountWei,
					})
					api.PrintResponse(response, err)
					return nil

				},
			},
			{
				Name:      "begin-reduce-bond-amount",
				Usage:     "Begin the bond reduction process for all provided minipools",
				UsageText: "rocketpool api minipool begin-reduce-bond-amount minipool-addresses new-bond-amount-wei",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}
					minipoolAddresses, err := cliutils.ValidateAddresses("minipool addresses", c.Args().Get(0))
					if err != nil {
						return err
					}
					newBondAmountWei, err := cliutils.ValidateWeiAmount("new bond amount", c.Args().Get(1))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(beginReduceBondAmounts(c, minipoolAddresses, newBondAmountWei))
					return nil

				},
			},

			// Change withdrawal creds
			{
				Name:      "can-change-withdrawal-creds",
				Usage:     "Check whether a solo validator's withdrawal credentials can be changed to a minipool address",
				UsageText: "rocketpool api minipool can-change-withdrawal-creds minipool-address mnemonic",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}
					minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
					if err != nil {
						return err
					}
					mnemonic, err := cliutils.ValidateWalletMnemonic("mnemonic", c.Args().Get(1))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canChangeWithdrawalCreds(c, minipoolAddress, mnemonic))
					return nil

				},
			},
			{
				Name:      "change-withdrawal-creds",
				Usage:     "Change a solo validator's withdrawal credentials to a minipool address",
				UsageText: "rocketpool api minipool change-withdrawal-creds minipool-address mnemonic",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}
					minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
					if err != nil {
						return err
					}
					mnemonic, err := cliutils.ValidateWalletMnemonic("mnemonic", c.Args().Get(1))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(changeWithdrawalCreds(c, minipoolAddress, mnemonic))
					return nil

				},
			},

			// Close
			{
				Name:      "get-minipool-close-details",
				Usage:     "Check all of the node's minipools for closure eligibility, and return the details of the closeable ones",
				UsageText: "rocketpool api minipool get-minipool-close-details",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					response, err := runMinipoolQuery[types.MinipoolCloseDetailsData](c, &minipoolCloseManager{})
					api.PrintResponse(response, err)
					return nil

				},
			},
			{
				Name:      "close",
				Aliases:   []string{"c"},
				Usage:     "Withdraw the balance from the specified dissolved minipools and close them",
				UsageText: "rocketpool api minipool close minipool-addresses",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					minipoolAddresses, err := cliutils.ValidateAddresses("minipool addresses", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(closeMinipools(c, minipoolAddresses))
					return nil

				},
			},

			// Delegate
			{
				Name:      "get-minipool-delegate-details",
				Usage:     "Get delegate information for all minipools belonging to the node",
				UsageText: "rocketpool api minipool get-minipool-delegate-details",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					response, err := runMinipoolQuery[types.MinipoolDelegateDetailsData](c, &minipoolDelegateManager{})
					api.PrintResponse(response, err)
					return nil

				},
			},
			{
				Name:      "upgrade-delegates",
				Usage:     "Upgrade the specified minipools to the latest network delegate contract",
				UsageText: "rocketpool api minipool upgrade-delegates minipool-addresses",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					minipoolAddresses, err := cliutils.ValidateAddresses("minipool addresses", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(upgradeDelegates(c, minipoolAddresses))
					return nil

				},
			},
			{
				Name:      "rollback-delegates",
				Usage:     "Rollback the specified minipools to their previous delegate contracts",
				UsageText: "rocketpool api minipool rollback-delegates minipool-addresses",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					minipoolAddresses, err := cliutils.ValidateAddresses("minipool addresses", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(rollbackDelegates(c, minipoolAddresses))
					return nil

				},
			},
			{
				Name:      "set-use-latest-delegates",
				Usage:     "Set whether or not to ignore the specified minipools's current delegate and always use the latest delegate instead",
				UsageText: "rocketpool api minipool set-use-latest-delegates minipool-addresses setting",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}
					minipoolAddresses, err := cliutils.ValidateAddresses("minipool addresses", c.Args().Get(0))
					if err != nil {
						return err
					}
					setting, err := cliutils.ValidateBool("setting", c.Args().Get(1))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(setUseLatestDelegates(c, minipoolAddresses, setting))
					return nil

				},
			},

			//  Dissolve
			{
				Name:      "get-minipool-dissolve-details",
				Usage:     "Get all of the details for dissolve eligibility of each node's minipools",
				UsageText: "rocketpool api minipool get-minipool-dissolve-details",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					response, err := runMinipoolQuery[types.MinipoolDissolveDetailsData](c, &minipoolDissolveManager{})
					api.PrintResponse(response, err)
					return nil

				},
			},
			{
				Name:      "dissolve",
				Aliases:   []string{"d"},
				Usage:     "Dissolve the specified initialized or prelaunch minipools",
				UsageText: "rocketpool api minipool dissolve minipool-addresses",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					minipoolAddresses, err := cliutils.ValidateAddresses("minipool addresses", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(dissolveMinipools(c, minipoolAddresses))
					return nil

				},
			},

			// Distribute
			{
				Name:      "get-distribute-balance-details",
				Usage:     "Get the balance distribution details for all of the node's minipools",
				UsageText: "rocketpool api minipool get-distribute-balance-details",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					response, err := runMinipoolQuery[types.MinipoolDistributeDetailsData](c, &minipoolDistributeManager{})
					api.PrintResponse(response, err)
					return nil

				},
			},
			{
				Name:      "distribute-balances",
				Usage:     "Distribute the specified minipools's ETH balances",
				UsageText: "rocketpool api minipool distribute-balance minipool-addresses",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					minipoolAddresses, err := cliutils.ValidateAddresses("minipool addresseses", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(distributeBalances(c, minipoolAddresses))
					return nil

				},
			},

			// Exit
			{
				Name:      "get-exit-details",
				Usage:     "Check whether any of the node's minipools can exit the Beacon chain",
				UsageText: "rocketpool api minipool get-exit-details",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					response, err := runMinipoolQuery[types.MinipoolExitDetailsData](c, &minipoolExitManager{})
					api.PrintResponse(response, err)
					return nil

				},
			},
			{
				Name:      "exit",
				Aliases:   []string{"e"},
				Usage:     "Exit the specified staking minipools from the Beacon chain",
				UsageText: "rocketpool api minipool exit minipool-addresses",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					minipoolAddresses, err := cliutils.ValidateAddresses("minipool addresses", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(exitMinipools(c, minipoolAddresses))
					return nil

				},
			},

			// Import key
			{
				Name:      "import-key",
				Usage:     "Import a validator private key for a vacant minipool",
				UsageText: "rocketpool api minipool import-key minipool-address mnemonic",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}
					minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
					if err != nil {
						return err
					}
					mnemonic, err := cliutils.ValidateWalletMnemonic("mnemonic", c.Args().Get(1))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(importKey(c, minipoolAddress, mnemonic))
					return nil

				},
			},

			// Promote
			{
				Name:      "get-promote-details",
				Usage:     "Check if any of the node's minipools are ready to be promoted and get their details",
				UsageText: "rocketpool api minipool get-promote-details",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					response, err := runMinipoolQuery[types.MinipoolPromoteDetailsData](c, &minipoolPromoteManager{})
					api.PrintResponse(response, err)
					return nil

				},
			},
			{
				Name:      "promote",
				Usage:     "Promote the specified vacant minipools",
				UsageText: "rocketpool api minipool promote minipool-addresses",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					minipoolAddresses, err := cliutils.ValidateAddresses("minipool addresses", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(promoteMinipools(c, minipoolAddresses))
					return nil

				},
			},

			// Reduce Bond
			{
				Name:      "get-reduce-bond-details",
				Usage:     "Check if any of the node's minipools are ready for bond reduction",
				UsageText: "rocketpool api minipool can-reduce-bond-amount minipool-address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					response, err := runMinipoolQuery[types.MinipoolReduceBondDetailsData](c, &minipoolReduceBondManager{})
					api.PrintResponse(response, err)
					return nil

				},
			},
			{
				Name:      "reduce-bond-amounts",
				Usage:     "Reduce the specified minipools's bonds",
				UsageText: "rocketpool api minipool reduce-bond-amounts minipool-addresses",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					minipoolAddresses, err := cliutils.ValidateAddresses("minipool addresses", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(reduceBondAmounts(c, minipoolAddresses))
					return nil

				},
			},

			// Refund
			{
				Name:      "get-refund-details",
				Usage:     "Get information about any available refunds for the node's minipools",
				UsageText: "rocketpool api minipool get-refund-details",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					response, err := runMinipoolQuery[types.MinipoolRefundDetailsData](c, &minipoolRefundManager{})
					api.PrintResponse(response, err)
					return nil

				},
			},
			{
				Name:      "refund",
				Aliases:   []string{"r"},
				Usage:     "Refund ETH belonging to the specified minipools",
				UsageText: "rocketpool api minipool refund minipool-addresses",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					minipoolAddresses, err := cliutils.ValidateAddresses("minipool addresses", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(refundMinipools(c, minipoolAddresses))
					return nil

				},
			},

			// Rescue Dissolved
			{
				Name:      "get-rescue-dissolved-details",
				Usage:     "Check all of the node's minipools for rescue eligibility, and return the details of the rescuable ones",
				UsageText: "rocketpool api minipool get-rescue-dissolved-details",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					response, err := runMinipoolQuery[types.MinipoolRescueDissolvedDetailsData](c, &minipoolRescueDissolvedManager{})
					api.PrintResponse(response, err)
					return nil

				},
			},
			{
				Name:      "rescue-dissolved",
				Usage:     "Rescue the specified dissolved minipools by depositing ETH for them to the Beacon deposit contract",
				UsageText: "rocketpool api minipool rescue-dissolved minipool-addresses deposit-amounts",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}
					minipoolAddresses, err := cliutils.ValidateAddresses("minipool addresses", c.Args().Get(0))
					if err != nil {
						return err
					}
					depositAmounts, err := cliutils.ValidateBigInts("deposit amounts", c.Args().Get(1))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(rescueDissolvedMinipools(c, minipoolAddresses, depositAmounts))
					return nil

				},
			},

			// Stake
			{
				Name:      "get-stake-details",
				Usage:     "Check whether any of the node's minipool are ready to be staked, moving from prelaunch to staking status",
				UsageText: "rocketpool api minipool get-stake-details",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					response, err := runMinipoolQuery[types.MinipoolStakeDetailsData](c, &minipoolStakeManager{})
					api.PrintResponse(response, err)
					return nil

				},
			},
			{
				Name:      "stake",
				Aliases:   []string{"t"},
				Usage:     "Stake the specified minipools, moving them from prelaunch to staking status",
				UsageText: "rocketpool api minipool stake minipool-address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					minipoolAddresses, err := cliutils.ValidateAddresses("minipool addresses", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(stakeMinipools(c, minipoolAddresses))
					return nil

				},
			},

			// Status
			{
				Name:      "status",
				Aliases:   []string{"s"},
				Usage:     "Get a list of the node's minipools",
				UsageText: "rocketpool api minipool status",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					response, err := runMinipoolQuery[types.MinipoolStatusData](c, &minipoolStatusManager{})
					api.PrintResponse(response, err)
					return nil

				},
			},

			// Vanity
			{
				Name:      "get-vanity-artifacts",
				Aliases:   []string{"v"},
				Usage:     "Gets the data necessary to search for vanity minipool addresses",
				UsageText: "rocketpool api minipool get-vanity-artifacts node-address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					nodeAddressStr := c.Args().Get(0)

					// Run
					api.PrintResponse(getVanityArtifacts(c, nodeAddressStr))
					return nil

				},
			},
		},
	})
}
