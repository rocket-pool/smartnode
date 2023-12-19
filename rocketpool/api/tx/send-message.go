package tx

import (
	"errors"
	"fmt"
	"net/url"
	_ "time/tzdata"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"

	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
)

// ===============
// === Factory ===
// ===============

type txSendMessageContextFactory struct {
	handler *TxHandler
}

func (f *txSendMessageContextFactory) Create(args url.Values) (*txSendMessageContext, error) {
	c := &txSendMessageContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("message", args, input.ValidateByteArray, &c.message),
		server.ValidateArg("address", args, input.ValidateAddress, &c.address),
	}
	return c, errors.Join(inputErrs...)
}

func (f *txSendMessageContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*txSendMessageContext, api.TxInfoData](
		router, "send-message", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type txSendMessageContext struct {
	handler *TxHandler
	message []byte
	address common.Address
}

func (c *txSendMessageContext) PrepareData(data *api.TxInfoData, opts *bind.TransactOpts) error {
	sp := c.handler.serviceProvider
	ec := sp.GetEthClient()

	err := errors.Join(
		sp.RequireWalletReady(),
	)
	if err != nil {
		return err
	}

	txInfo, err := core.NewTransactionInfoRaw(ec, c.address, c.message, opts)
	if err != nil {
		return fmt.Errorf("error creating TX info: %w", err)
	}
	data.TxInfo = txInfo
	return nil
}
