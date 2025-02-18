package megapool

import (
	"context"

	"github.com/gorilla/mux"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/services"
)

type MegapoolHandler struct {
	logger          *log.Logger
	ctx             context.Context
	serviceProvider *services.ServiceProvider
	factories       []server.IContextFactory
}

func NewMegapoolHandler(logger *log.Logger, ctx context.Context, serviceProvider *services.ServiceProvider) *MegapoolHandler {
	h := &MegapoolHandler{
		logger:          logger,
		ctx:             ctx,
		serviceProvider: serviceProvider,
	}
	h.factories = []server.IContextFactory{
		&megapoolStatusContextFactory{h},
	}
	return h
}

func (h *MegapoolHandler) RegisterRoutes(router *mux.Router) {
	subrouter := router.PathPrefix("/megapool").Subrouter()
	for _, factory := range h.factories {
		factory.RegisterRoute(subrouter)
	}
}
