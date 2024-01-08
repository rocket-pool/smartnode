package tx

import (
	"github.com/gorilla/mux"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/rocketpool/common/services"
)

type TxHandler struct {
	serviceProvider *services.ServiceProvider
	factories       []server.IContextFactory
}

func NewTxHandler(serviceProvider *services.ServiceProvider) *TxHandler {
	h := &TxHandler{
		serviceProvider: serviceProvider,
	}
	h.factories = []server.IContextFactory{
		&txBatchSignTxsContextFactory{h},
		&txBatchSubmitTxsContextFactory{h},
		&txSendMessageContextFactory{h},
		&txSignMessageContextFactory{h},
		&txSignTxContextFactory{h},
		&txSubmitTxContextFactory{h},
		&txWaitContextFactory{h},
	}
	return h
}

func (h *TxHandler) RegisterRoutes(router *mux.Router) {
	subrouter := router.PathPrefix("/tx").Subrouter()
	for _, factory := range h.factories {
		factory.RegisterRoute(subrouter)
	}
}
