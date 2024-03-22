package minipool

import (
	"github.com/gorilla/mux"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/services"
)

type MinipoolHandler struct {
	serviceProvider *services.ServiceProvider
	factories       []server.IContextFactory
}

func NewMinipoolHandler(serviceProvider *services.ServiceProvider) *MinipoolHandler {
	h := &MinipoolHandler{
		serviceProvider: serviceProvider,
	}
	h.factories = []server.IContextFactory{
		&minipoolBeginReduceBondDetailsContextFactory{h},
		&minipoolBeginReduceBondContextFactory{h},
		&minipoolCanChangeCredsContextFactory{h},
		&minipoolChangeCredsContextFactory{h},
		&minipoolCloseDetailsContextFactory{h},
		&minipoolCloseContextFactory{h},
		&minipoolDelegateDetailsContextFactory{h},
		&minipoolUpgradeDelegatesContextFactory{h},
		&minipoolRollbackDelegatesContextFactory{h},
		&minipoolSetUseLatestDelegatesContextFactory{h},
		&minipoolDissolveDetailsContextFactory{h},
		&minipoolDissolveContextFactory{h},
		&minipoolDistributeDetailsContextFactory{h},
		&minipoolDistributeContextFactory{h},
		&minipoolExitDetailsContextFactory{h},
		&minipoolExitContextFactory{h},
		&minipoolImportKeyContextFactory{h},
		&minipoolPromoteDetailsContextFactory{h},
		&minipoolPromoteContextFactory{h},
		&minipoolReduceBondDetailsContextFactory{h},
		&minipoolReduceBondContextFactory{h},
		&minipoolRefundDetailsContextFactory{h},
		&minipoolRefundContextFactory{h},
		&minipoolRescueDissolvedDetailsContextFactory{h},
		&minipoolRescueDissolvedContextFactory{h},
		&minipoolStakeDetailsContextFactory{h},
		&minipoolStakeContextFactory{h},
		&minipoolStatusContextFactory{h},
		&minipoolVanityContextFactory{h},
	}
	return h
}

func (h *MinipoolHandler) RegisterRoutes(router *mux.Router) {
	subrouter := router.PathPrefix("/minipool").Subrouter()
	for _, factory := range h.factories {
		factory.RegisterRoute(subrouter)
	}
}
