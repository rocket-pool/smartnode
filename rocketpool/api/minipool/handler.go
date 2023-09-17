package minipool

import (
	"github.com/gorilla/mux"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/rocketpool/common/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
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
	reduceBondDetailsFactory      server.IMinipoolCallContextFactory[*minipoolReduceBondDetailsContext, api.MinipoolReduceBondDetailsData]
	reduceBondFactory             server.IQuerylessCallContextFactory[*minipoolReduceBondContext, api.BatchTxInfoData]
	refundDetailsFactory          server.IMinipoolCallContextFactory[*minipoolRefundDetailsContext, api.MinipoolRefundDetailsData]
	refundFactory                 server.IQuerylessCallContextFactory[*minipoolRefundContext, api.BatchTxInfoData]
	rescueDissolvedDetailsFactory server.IMinipoolCallContextFactory[*minipoolRescueDissolvedDetailsContext, api.MinipoolRescueDissolvedDetailsData]
	rescueDissolvedFactory        server.IQuerylessCallContextFactory[*minipoolRescueDissolvedContext, api.BatchTxInfoData]
	stakeDetailsFactory           server.IMinipoolCallContextFactory[*minipoolStakeDetailsContext, api.MinipoolStakeDetailsData]
	stakeFactory                  server.IQuerylessCallContextFactory[*minipoolStakeContext, api.BatchTxInfoData]
	statusFactory                 server.IMinipoolCallContextFactory[*minipoolStatusContext, api.MinipoolStatusData]
	vanityFactory                 server.IQuerylessCallContextFactory[*minipoolVanityContext, api.MinipoolVanityArtifactsData]
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
	h.reduceBondDetailsFactory = &minipoolReduceBondDetailsContextFactory{h}
	h.reduceBondFactory = &minipoolReduceBondContextFactory{h}
	h.refundDetailsFactory = &minipoolRefundDetailsContextFactory{h}
	h.refundFactory = &minipoolRefundContextFactory{h}
	h.rescueDissolvedDetailsFactory = &minipoolRescueDissolvedDetailsContextFactory{h}
	h.rescueDissolvedFactory = &minipoolRescueDissolvedContextFactory{h}
	h.stakeDetailsFactory = &minipoolStakeDetailsContextFactory{h}
	h.stakeFactory = &minipoolStakeContextFactory{h}
	h.statusFactory = &minipoolStatusContextFactory{h}
	h.vanityFactory = &minipoolVanityContextFactory{h}
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
	server.RegisterMinipoolRoute(router, "reduce-bond/details", h.reduceBondDetailsFactory, h.serviceProvider)
	server.RegisterQuerylessRoute(router, "reduce-bond", h.reduceBondFactory, h.serviceProvider)
	server.RegisterMinipoolRoute(router, "refund/details", h.refundDetailsFactory, h.serviceProvider)
	server.RegisterQuerylessRoute(router, "refund", h.refundFactory, h.serviceProvider)
	server.RegisterMinipoolRoute(router, "rescue-dissolved/details", h.rescueDissolvedDetailsFactory, h.serviceProvider)
	server.RegisterQuerylessRoute(router, "rescue-dissolved", h.rescueDissolvedFactory, h.serviceProvider)
	server.RegisterMinipoolRoute(router, "stake/details", h.stakeDetailsFactory, h.serviceProvider)
	server.RegisterQuerylessRoute(router, "stake", h.stakeFactory, h.serviceProvider)
	server.RegisterMinipoolRoute(router, "status", h.statusFactory, h.serviceProvider)
	server.RegisterQuerylessRoute(router, "vanity-artifacts", h.vanityFactory, h.serviceProvider)
}
