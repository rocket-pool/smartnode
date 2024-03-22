package queue

import (
	"github.com/gorilla/mux"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/services"
)

type QueueHandler struct {
	serviceProvider *services.ServiceProvider
	factories       []server.IContextFactory
}

func NewQueueHandler(serviceProvider *services.ServiceProvider) *QueueHandler {
	h := &QueueHandler{
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
