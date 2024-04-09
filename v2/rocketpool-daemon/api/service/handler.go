package service

import (
	"context"

	"github.com/gorilla/mux"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/services"
)

type ServiceHandler struct {
	logger          *log.Logger
	ctx             context.Context
	serviceProvider *services.ServiceProvider
	factories       []server.IContextFactory
}

func NewServiceHandler(logger *log.Logger, ctx context.Context, serviceProvider *services.ServiceProvider) *ServiceHandler {
	h := &ServiceHandler{
		logger:          logger,
		ctx:             ctx,
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
