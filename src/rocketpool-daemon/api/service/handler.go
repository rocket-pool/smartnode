package service

import (
	"github.com/gorilla/mux"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/services"
)

type ServiceHandler struct {
	serviceProvider *services.ServiceProvider
	factories       []server.IContextFactory
}

func NewServiceHandler(serviceProvider *services.ServiceProvider) *ServiceHandler {
	h := &ServiceHandler{
		serviceProvider: serviceProvider,
	}
	h.factories = []server.IContextFactory{
		&serviceClientStatusContextFactory{h},
		&serviceRestartVcContextFactory{h},
		&serviceTerminateDataFolderContextFactory{h},
		&serviceVersionContextFactory{h},
	}
	return h
}

func (h *ServiceHandler) RegisterRoutes(router *mux.Router) {
	subrouter := router.PathPrefix("/service").Subrouter()
	for _, factory := range h.factories {
		factory.RegisterRoute(subrouter)
	}
}
