package minipool

import (
	"errors"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
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
		server.ValidateArg("addresses", args, input.ValidateAddresses, &c.minipoolAddresses),
	}
	return c, errors.Join(inputErrs...)
}

func (f *minipoolRefundContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*minipoolRefundContext, api.BatchTxInfoData](
		router, "refund", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolRefundContext struct {
	handler           *MinipoolHandler
	minipoolAddresses []common.Address
}

func (c *minipoolRefundContext) PrepareData(data *api.BatchTxInfoData, opts *bind.TransactOpts) error {
	return prepareMinipoolBatchTxData(c.handler.serviceProvider, c.minipoolAddresses, data, c.CreateTx, "refund")
}

func (c *minipoolRefundContext) CreateTx(mp minipool.IMinipool, opts *bind.TransactOpts) (*core.TransactionInfo, error) {
	return mp.Common().Refund(opts)
}
