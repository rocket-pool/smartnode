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
	beginReduceBondFactory        server.IQuerylessCallContextFactory[*minipoolBeginReduceBondContext, api.BatchTxInfoData]
	canChangeCredsFactory         server.ISingleStageCallContextFactory[*minipoolCanChangeCredsContext, api.MinipoolCanChangeWithdrawalCredentialsData]
	changeCredsFactory            server.ISingleStageCallContextFactory[*minipoolChangeCredsContext, api.SuccessData]
	closeDetailsFactory           server.IMinipoolCallContextFactory[*minipoolCloseDetailsContext, api.MinipoolCloseDetailsData]
	closeFactory                  server.IMinipoolCallContextFactory[*minipoolCloseContext, api.BatchTxInfoData]
	delegateDetailsFactory        server.IMinipoolCallContextFactory[*minipoolDelegateDetailsContext, api.MinipoolDelegateDetailsData]
	upgradeDelegatesFactory       server.IQuerylessCallContextFactory[*minipoolUpgradeDelegatesContext, api.BatchTxInfoData]
	rollbackDelegatesFactory      server.IQuerylessCallContextFactory[*minipoolRollbackDelegatesContext, api.BatchTxInfoData]
	dissolveDetailsFactory        server.IMinipoolCallContextFactory[*minipoolDissolveDetailsContext, api.MinipoolDissolveDetailsData]
	dissolveFactory               server.IQuerylessCallContextFactory[*minipoolDissolveContext, api.BatchTxInfoData]
	distributeDetailsFactory      server.IMinipoolCallContextFactory[*minipoolDistributeDetailsContext, api.MinipoolDistributeDetailsData]
	distributeFactory             server.IQuerylessCallContextFactory[*minipoolDistributeContext, api.BatchTxInfoData]
	exitDetailsFactory            server.IMinipoolCallContextFactory[*minipoolExitDetailsContext, api.MinipoolExitDetailsData]
	exitFactory                   server.IMinipoolCallContextFactory[*minipoolExitContext, api.SuccessData]
	importFactory                 server.ISingleStageCallContextFactory[*minipoolImportKeyContext, api.SuccessData]
	promoteDetailsFactory         server.IMinipoolCallContextFactory[*minipoolPromoteDetailsContext, api.MinipoolPromoteDetailsData]
	promoteFactory                server.IQuerylessCallContextFactory[*minipoolPromoteContext, api.BatchTxInfoData]
}

func NewMinipoolHandler(serviceProvider *services.ServiceProvider) *MinipoolHandler {
	h := &MinipoolHandler{
		serviceProvider: serviceProvider,
	}
	h.beginReduceBondDetailsFactory = &minipoolBeginReduceBondDetailsContextFactory{h}
	h.beginReduceBondFactory = &minipoolBeginReduceBondContextFactory{h}
	h.canChangeCredsFactory = &minipoolCanChangeCredsContextFactory{h}
	h.changeCredsFactory = &minipoolChangeCredsContextFactory{h}
	h.closeDetailsFactory = &minipoolCloseDetailsContextFactory{h}
	h.closeFactory = &minipoolCloseContextFactory{h}
	h.delegateDetailsFactory = &minipoolDelegateDetailsContextFactory{h}
	h.upgradeDelegatesFactory = &minipoolUpgradeDelegatesContextFactory{h}
	h.rollbackDelegatesFactory = &minipoolRollbackDelegatesContextFactory{h}
	h.dissolveDetailsFactory = &minipoolDissolveDetailsContextFactory{h}
	h.dissolveFactory = &minipoolDissolveContextFactory{h}
	h.distributeDetailsFactory = &minipoolDistributeDetailsContextFactory{h}
	h.distributeFactory = &minipoolDistributeContextFactory{h}
	h.exitDetailsFactory = &minipoolExitDetailsContextFactory{h}
	h.exitFactory = &minipoolExitContextFactory{h}
	h.importFactory = &minipoolImportKeyContextFactory{h}
	h.promoteDetailsFactory = &minipoolPromoteDetailsContextFactory{h}
	h.promoteFactory = &minipoolPromoteContextFactory{h}
	return h
}

func (h *MinipoolHandler) RegisterRoutes(router *mux.Router) {
	server.RegisterMinipoolRoute(router, "begin-reduce-bond/details", h.beginReduceBondDetailsFactory, h.serviceProvider)
	server.RegisterQuerylessRoute(router, "begin-reduce-bond", h.beginReduceBondFactory, h.serviceProvider)
	server.RegisterSingleStageRoute(router, "change-withdrawal-creds/verify", h.canChangeCredsFactory, h.serviceProvider)
	server.RegisterSingleStageRoute(router, "change-withdrawal-creds", h.changeCredsFactory, h.serviceProvider)
	server.RegisterMinipoolRoute(router, "close/details", h.closeDetailsFactory, h.serviceProvider)
	server.RegisterMinipoolRoute(router, "close", h.closeFactory, h.serviceProvider)
	server.RegisterMinipoolRoute(router, "delegate/details", h.delegateDetailsFactory, h.serviceProvider)
	server.RegisterQuerylessRoute(router, "delegate/upgrade", h.upgradeDelegatesFactory, h.serviceProvider)
	server.RegisterQuerylessRoute(router, "delegate/rollback", h.rollbackDelegatesFactory, h.serviceProvider)
	server.RegisterMinipoolRoute(router, "dissolve/details", h.dissolveDetailsFactory, h.serviceProvider)
	server.RegisterQuerylessRoute(router, "dissolve", h.dissolveFactory, h.serviceProvider)
	server.RegisterMinipoolRoute(router, "distribute/details", h.distributeDetailsFactory, h.serviceProvider)
	server.RegisterQuerylessRoute(router, "distribute", h.distributeFactory, h.serviceProvider)
	server.RegisterMinipoolRoute(router, "exit/details", h.exitDetailsFactory, h.serviceProvider)
	server.RegisterMinipoolRoute(router, "exit", h.exitFactory, h.serviceProvider)
	server.RegisterSingleStageRoute(router, "import-key", h.importFactory, h.serviceProvider)
	server.RegisterMinipoolRoute(router, "promote/details", h.promoteDetailsFactory, h.serviceProvider)
	server.RegisterQuerylessRoute(router, "promote", h.promoteFactory, h.serviceProvider)
}

// Register subcommands
func RegisterSubcommands(command *cli.Command, name string, aliases []string) {
	command.Subcommands = append(command.Subcommands, cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Manage the node's minipools",
		Subcommands: []cli.Command{

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
