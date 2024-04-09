package security

import (
	"context"

	"github.com/gorilla/mux"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/services"
)

type SecurityCouncilHandler struct {
	logger          *log.Logger
	ctx             context.Context
	serviceProvider *services.ServiceProvider
	factories       []server.IContextFactory
}

func NewSecurityCouncilHandler(logger *log.Logger, ctx context.Context, serviceProvider *services.ServiceProvider) *SecurityCouncilHandler {
	h := &SecurityCouncilHandler{
		logger:          logger,
		ctx:             ctx,
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
