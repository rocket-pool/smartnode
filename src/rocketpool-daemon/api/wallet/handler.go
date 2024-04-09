package wallet

import (
	"context"

	"github.com/gorilla/mux"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/services"
)

type WalletHandler struct {
	logger          *log.Logger
	ctx             context.Context
	serviceProvider *services.ServiceProvider
	factories       []server.IContextFactory
}

func NewWalletHandler(logger *log.Logger, ctx context.Context, serviceProvider *services.ServiceProvider) *WalletHandler {
	h := &WalletHandler{
		logger:          logger,
		ctx:             ctx,
		serviceProvider: serviceProvider,
	}
	h.factories = []server.IContextFactory{
		&walletCreateValidatorKeyContextFactory{h},
		&walletDeletePasswordContextFactory{h},
		&walletExportContextFactory{h},
		&walletExportEthKeyContextFactory{h},
		&walletInitializeContextFactory{h},
		&walletMasqueradeContextFactory{h},
		&walletRebuildContextFactory{h},
		&walletRecoverContextFactory{h},
		&walletRestoreAddressContextFactory{h},
		&walletSearchAndRecoverContextFactory{h},
		&walletSendMessageContextFactory{h},
		&walletSetEnsNameContextFactory{h},
		&walletSetPasswordContextFactory{h},
		&walletSignMessageContextFactory{h},
		&walletSignTxContextFactory{h},
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
