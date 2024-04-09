package queue

import (
	"context"

	"github.com/gorilla/mux"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/services"
)

type QueueHandler struct {
	logger          *log.Logger
	ctx             context.Context
	serviceProvider *services.ServiceProvider
	factories       []server.IContextFactory
}

func NewQueueHandler(logger *log.Logger, ctx context.Context, serviceProvider *services.ServiceProvider) *QueueHandler {
	h := &QueueHandler{
		logger:          logger,
		ctx:             ctx,
		serviceProvider: serviceProvider,
	}
	h.factories = []server.IContextFactory{
		&queueProcessContextFactory{h},
		&queueStatusContextFactory{h},
	}
	return h
}

func (h *QueueHandler) RegisterRoutes(router *mux.Router) {
	subrouter := router.PathPrefix("/queue").Subrouter()
	for _, factory := range h.factories {
		factory.RegisterRoute(subrouter)
	}
}
