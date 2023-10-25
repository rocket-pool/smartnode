package service

import (
	"github.com/gorilla/mux"
	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/rocketpool/common/services"
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
	}
	return h
}

func (h *ServiceHandler) RegisterRoutes(router *mux.Router) {
	for _, factory := range h.factories {
		factory.RegisterRoute(router)
	}
}
