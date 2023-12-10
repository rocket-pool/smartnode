package pdao

import (
	"github.com/gorilla/mux"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/rocketpool/common/services"
)

type ProtocolDaoHandler struct {
	serviceProvider *services.ServiceProvider
	factories       []server.IContextFactory
}

func NewProtocolDaoHandler(serviceProvider *services.ServiceProvider) *ProtocolDaoHandler {
	h := &ProtocolDaoHandler{
		serviceProvider: serviceProvider,
	}
	h.factories = []server.IContextFactory{
		&protocolDaoClaimBondsContextFactory{h},
		&protocolDaoDefeatProposalContextFactory{h},
		&protocolDaoExecuteProposalContextFactory{h},
		&protocolDaoFinalizeProposalContextFactory{h},
		&protocolDaoGetClaimableBondsContextFactory{h},
		&protocolDaoProposeInviteToSecurityCouncilContextFactory{h},
		&protocolDaoSettingsContextFactory{h},
	}
	return h
}

func (h *ProtocolDaoHandler) RegisterRoutes(router *mux.Router) {
	subrouter := router.PathPrefix("/pdao").Subrouter()
	for _, factory := range h.factories {
		factory.RegisterRoute(subrouter)
	}
}
