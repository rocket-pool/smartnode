package security

import (
	"github.com/gorilla/mux"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/services"
)

type SecurityCouncilHandler struct {
	serviceProvider *services.ServiceProvider
	factories       []server.IContextFactory
}

func NewSecurityCouncilHandler(serviceProvider *services.ServiceProvider) *SecurityCouncilHandler {
	h := &SecurityCouncilHandler{
		serviceProvider: serviceProvider,
	}
	h.factories = []server.IContextFactory{
		&securityCancelProposalContextFactory{h},
		&securityExecuteProposalsContextFactory{h},
		&securityJoinContextFactory{h},
		&securityLeaveContextFactory{h},
		&securityMembersContextFactory{h},
		&securityProposalsContextFactory{h},
		&securityProposeLeaveContextFactory{h},
		&securityProposeSettingContextFactory{h},
		&securityStatusContextFactory{h},
		&securityVoteOnProposalContextFactory{h},
	}
	return h
}

func (h *SecurityCouncilHandler) RegisterRoutes(router *mux.Router) {
	subrouter := router.PathPrefix("/security").Subrouter()
	for _, factory := range h.factories {
		factory.RegisterRoute(subrouter)
	}
}
