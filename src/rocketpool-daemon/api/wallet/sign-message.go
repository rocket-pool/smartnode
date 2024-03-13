package wallet

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	_ "time/tzdata"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/shared/types/api"
	hexutils "github.com/rocket-pool/smartnode/shared/utils/hex"
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
		router, "sign-message", f, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type walletSignMessageContext struct {
	handler *WalletHandler
	message []byte
}

func (c *walletSignMessageContext) PrepareData(data *api.WalletSignMessageData, opts *bind.TransactOpts) error {
	sp := c.handler.serviceProvider
	w := sp.GetWallet()

	err := errors.Join(
		sp.RequireWalletReady(),
	)
	if err != nil {
		return err
	}

	signedBytes, err := w.SignMessage(c.message)
	if err != nil {
		return fmt.Errorf("error signing message: %w", err)
	}
	data.SignedMessage = hexutils.AddPrefix(hex.EncodeToString(signedBytes))
	return nil
}
