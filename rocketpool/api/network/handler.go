package network

import (
	"github.com/gorilla/mux"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/rocketpool/common/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

type NetworkHandler struct {
	serviceProvider        *services.ServiceProvider
	proposalsFactory       server.ISingleStageCallContextFactory[*networkProposalContext, api.NetworkDaoProposalsData]
	delegateFactory        server.ISingleStageCallContextFactory[*networkDelegateContext, api.NetworkLatestDelegateData]
	depositInfoFactory     server.ISingleStageCallContextFactory[*networkDepositInfoContext, api.NetworkDepositContractInfoData]
	downloadRewardsFactory server.ISingleStageCallContextFactory[*networkDownloadRewardsContext, api.SuccessData]
	rewardsFileFactory     server.ISingleStageCallContextFactory[*networkRewardsFileContext, api.NetworkRewardsFileData]
	generateRewardsFactory server.ISingleStageCallContextFactory[*networkGenerateRewardsContext, api.SuccessData]
	feeFactory             server.ISingleStageCallContextFactory[*networkFeeContext, api.NetworkNodeFeeData]
	priceFactory           server.ISingleStageCallContextFactory[*networkPriceContext, api.NetworkRplPriceData]
	statsFactory           server.ISingleStageCallContextFactory[*networkStatsContext, api.NetworkStatsData]
	timezoneFactory        server.ISingleStageCallContextFactory[*networkTimezoneContext, api.NetworkTimezonesData]
}

func NewNetworkHandler(serviceProvider *services.ServiceProvider) *NetworkHandler {
	h := &NetworkHandler{
		serviceProvider: serviceProvider,
	}
	h.proposalsFactory = &networkProposalContextFactory{h}
	h.delegateFactory = &networkDelegateContextFactory{h}
	h.depositInfoFactory = &networkDepositInfoContextFactory{h}
	h.downloadRewardsFactory = &networkDownloadRewardsContextFactory{h}
	h.rewardsFileFactory = &networkRewardsFileContextFactory{h}
	h.generateRewardsFactory = &networkGenerateRewardsContextFactory{h}
	h.feeFactory = &networkFeeContextFactory{h}
	h.priceFactory = &networkPriceContextFactory{h}
	h.statsFactory = &networkStatsContextFactory{h}
	h.timezoneFactory = &networkTimezoneContextFactory{h}
	return h
}

func (h *NetworkHandler) RegisterRoutes(router *mux.Router) {
	server.RegisterSingleStageRoute(router, "dao-proposals", h.proposalsFactory, h.serviceProvider)
	server.RegisterSingleStageRoute(router, "latest-delegate", h.delegateFactory, h.serviceProvider)
	server.RegisterSingleStageRoute(router, "deposit-contract-info", h.depositInfoFactory, h.serviceProvider)
	server.RegisterSingleStageRoute(router, "download-rewards-file", h.downloadRewardsFactory, h.serviceProvider)
	server.RegisterSingleStageRoute(router, "get-rewards-file-info", h.rewardsFileFactory, h.serviceProvider)
	server.RegisterSingleStageRoute(router, "generate-rewards-tree", h.generateRewardsFactory, h.serviceProvider)
	server.RegisterSingleStageRoute(router, "node-fee", h.feeFactory, h.serviceProvider)
	server.RegisterSingleStageRoute(router, "rpl-price", h.priceFactory, h.serviceProvider)
	server.RegisterSingleStageRoute(router, "stats", h.statsFactory, h.serviceProvider)
	server.RegisterSingleStageRoute(router, "timezone-map", h.timezoneFactory, h.serviceProvider)
}
