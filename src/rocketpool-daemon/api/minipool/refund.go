package minipool

import (
	"errors"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/rocketpool-go/minipool"
)

// ===============
// === Factory ===
// ===============

type minipoolRefundContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolRefundContextFactory) Create(args url.Values) (*minipoolRefundContext, error) {
	c := &minipoolRefundContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArgBatch("addresses", args, minipoolAddressBatchSize, input.ValidateAddress, &c.minipoolAddresses),
	}
	return c, errors.Join(inputErrs...)
}

func (f *minipoolRefundContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*minipoolRefundContext, types.BatchTxInfoData](
		router, "refund", f, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolRefundContext struct {
	handler           *MinipoolHandler
	minipoolAddresses []common.Address
}

func (c *minipoolRefundContext) PrepareData(data *types.BatchTxInfoData, opts *bind.TransactOpts) error {
	return prepareMinipoolBatchTxData(c.handler.serviceProvider, c.minipoolAddresses, data, c.CreateTx, "refund")
}

func (c *minipoolRefundContext) CreateTx(mp minipool.IMinipool, opts *bind.TransactOpts) (*eth.TransactionInfo, error) {
	return mp.Common().Refund(opts)
}
