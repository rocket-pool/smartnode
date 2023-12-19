package tx

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	_ "time/tzdata"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	hexutils "github.com/rocket-pool/smartnode/shared/utils/hex"
	"github.com/rocket-pool/smartnode/shared/utils/input"
)

// ===============
// === Factory ===
// ===============

type txSignMessageContextFactory struct {
	handler *TxHandler
}

func (f *txSignMessageContextFactory) Create(args url.Values) (*txSignMessageContext, error) {
	c := &txSignMessageContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("message", args, input.ValidateByteArray, &c.message),
	}
	return c, errors.Join(inputErrs...)
}

func (f *txSignMessageContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*txSignMessageContext, api.TxSignMessageData](
		router, "sign-message", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type txSignMessageContext struct {
	handler *TxHandler
	message []byte
}

func (c *txSignMessageContext) PrepareData(data *api.TxSignMessageData, opts *bind.TransactOpts) error {
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
