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
	}
	return h
}

func (h *SecurityCouncilHandler) RegisterRoutes(router *mux.Router) {
	subrouter := router.PathPrefix("/security").Subrouter()
	for _, factory := range h.factories {
		factory.RegisterRoute(subrouter)
	}
}
