package pdao

import (
	"context"

	"github.com/gorilla/mux"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/services"
)

type ProtocolDaoHandler struct {
	logger          *log.Logger
	ctx             context.Context
	serviceProvider *services.ServiceProvider
	factories       []server.IContextFactory
}

func NewProtocolDaoHandler(logger *log.Logger, ctx context.Context, serviceProvider *services.ServiceProvider) *ProtocolDaoHandler {
	h := &ProtocolDaoHandler{
		logger:          logger,
		ctx:             ctx,
		serviceProvider: serviceProvider,
	}
	h.factories = []server.IContextFactory{
		&protocolDaoClaimBondsContextFactory{h},
		&protocolDaoDefeatProposalContextFactory{h},
		&protocolDaoExecuteProposalsContextFactory{h},
		&protocolDaoFinalizeProposalContextFactory{h},
		&protocolDaoOverrideVoteOnProposalContextFactory{h},
		&protocolDaoVoteOnProposalContextFactory{h},
		&protocolDaoGetClaimableBondsContextFactory{h},
		&protocolDaoProposeOneTimeSpendContextFactory{h},
		&protocolDaoProposeRecurringSpendContextFactory{h},
		&protocolDaoProposeRecurringSpendUpdateContextFactory{h},
		&protocolDaoProposeInviteToSecurityCouncilContextFactory{h},
		&protocolDaoProposeKickFromSecurityCouncilContextFactory{h},
		&protocolDaoProposeKickMultiFromSecurityCouncilContextFactory{h},
		&protocolDaoProposeReplaceMemberOfSecurityCouncilContextFactory{h},
		&protocolDaoProposalsContextFactory{h},
		&protocolDaoRewardsPercentagesContextFactory{h},
		&protocolDaoProposeRewardsPercentagesContextFactory{h},
		&protocolDaoSettingsContextFactory{h},
		&protocolDaoProposeSettingContextFactory{h},
		&protocolDaoInitializeVotingContextFactory{h},
		&protocolDaoSetVotingDelegateContextFactory{h},
		&protocolDaoCurrentVotingDelegateContextFactory{h},
	}
	return h
}

func (h *ProtocolDaoHandler) RegisterRoutes(router *mux.Router) {
	subrouter := router.PathPrefix("/pdao").Subrouter()
	for _, factory := range h.factories {
		factory.RegisterRoute(subrouter)
	}
}
