package wallet

import (
	"github.com/gorilla/mux"

	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/server"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/services"
)

type WalletHandler struct {
	serviceProvider *services.ServiceProvider
	factories       []server.IContextFactory
}

func NewWalletHandler(serviceProvider *services.ServiceProvider) *WalletHandler {
	h := &WalletHandler{
		serviceProvider: serviceProvider,
	}
	h.factories = []server.IContextFactory{
		&walletCreateValidatorKeyContextFactory{h},
		&walletDeletePasswordContextFactory{h},
		&walletExportContextFactory{h},
		&walletForgetPasswordContextFactory{h},
		&walletInitializeContextFactory{h},
		&walletRebuildContextFactory{h},
		&walletRecoverContextFactory{h},
		&walletSavePasswordContextFactory{h},
		&walletSearchAndRecoverContextFactory{h},
		&walletSendMessageContextFactory{h},
		&walletSetEnsNameContextFactory{h},
		&walletSetPasswordContextFactory{h},
		&walletSignMessageContextFactory{h},
		&walletStatusFactory{h},
		&walletTestRecoverContextFactory{h},
		&walletTestSearchAndRecoverContextFactory{h},
	}
	return h
}

func (h *WalletHandler) RegisterRoutes(router *mux.Router) {
	subrouter := router.PathPrefix("/wallet").Subrouter()
	for _, factory := range h.factories {
		factory.RegisterRoute(subrouter)
	}
}
