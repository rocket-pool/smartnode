package wallet

import (
	"errors"
	"net/url"
	_ "time/tzdata"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/utils/input"
)

// ===============
// === Factory ===
// ===============

type walletSendMessageContextFactory struct {
	handler *WalletHandler
}

func (f *walletSendMessageContextFactory) Create(args url.Values) (*walletSendMessageContext, error) {
	c := &walletSendMessageContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("message", args, input.ValidateByteArray, &c.message),
		server.ValidateArg("address", args, input.ValidateAddress, &c.address),
	}
	return c, errors.Join(inputErrs...)
}

func (f *walletSendMessageContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*walletSendMessageContext, types.TxInfoData](
		router, "send-message", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type walletSendMessageContext struct {
	handler *WalletHandler
	message []byte
	address common.Address
}

func (c *walletSendMessageContext) PrepareData(data *types.TxInfoData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	txMgr := sp.GetTransactionManager()

	// Requirements
	err := sp.RequireNodeAddress()
	if err != nil {
		return types.ResponseStatus_AddressNotPresent, err
	}

	txInfo := txMgr.CreateTransactionInfoRaw(c.address, c.message, opts)
	data.TxInfo = txInfo
	return types.ResponseStatus_Success, nil
}
