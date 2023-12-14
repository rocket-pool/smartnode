package security

import (
	"github.com/gorilla/mux"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/rocketpool/common/services"
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
		&securityExecuteProposalContextFactory{h},
		&securityJoinContextFactory{h},
		&securityLeaveContextFactory{h},
		&securityMembersContextFactory{h},
		&securityProposalsContextFactory{h},
		&securityProposeInviteContextFactory{h},
		&securityProposeLeaveContextFactory{h},
		&securityProposeKickContextFactory{h},
		&securityProposeKickMultiContextFactory{h},
		&securityProposeReplaceContextFactory{h},
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
