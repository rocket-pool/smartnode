package wallet

import (
	"errors"
	"fmt"
	"net/url"
	_ "time/tzdata"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type walletSignMessageContextFactory struct {
	handler *WalletHandler
}

func (f *walletSignMessageContextFactory) Create(args url.Values) (*walletSignMessageContext, error) {
	c := &walletSignMessageContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("message", args, input.ValidateByteArray, &c.message),
	}
	return c, errors.Join(inputErrs...)
}

func (f *walletSignMessageContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*walletSignMessageContext, api.WalletSignMessageData](
		router, "sign-message", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type walletSignMessageContext struct {
	handler *WalletHandler
	message []byte
}

func (c *walletSignMessageContext) PrepareData(data *api.WalletSignMessageData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	w := sp.GetWallet()

	// Requirements
	err := sp.RequireWalletReady()
	if err != nil {
		return types.ResponseStatus_WalletNotReady, err
	}

	signedBytes, err := w.SignMessage(c.message)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error signing message: %w", err)
	}
	data.SignedMessage = signedBytes
	return types.ResponseStatus_Success, nil
}
