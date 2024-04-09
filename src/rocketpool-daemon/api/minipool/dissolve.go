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
	"github.com/rocket-pool/rocketpool-go/v2/minipool"
)

// ===============
// === Factory ===
// ===============

type minipoolDissolveContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolDissolveContextFactory) Create(args url.Values) (*minipoolDissolveContext, error) {
	c := &minipoolDissolveContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArgBatch("addresses", args, minipoolAddressBatchSize, input.ValidateAddress, &c.minipoolAddresses),
	}
	return c, errors.Join(inputErrs...)
}

func (f *minipoolDissolveContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*minipoolDissolveContext, types.BatchTxInfoData](
		router, "dissolve", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolDissolveContext struct {
	handler           *MinipoolHandler
	minipoolAddresses []common.Address
}

func (c *minipoolDissolveContext) PrepareData(data *types.BatchTxInfoData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	return prepareMinipoolBatchTxData(c.handler.ctx, c.handler.serviceProvider, c.minipoolAddresses, data, c.CreateTx, "dissolve")
}

func (c *minipoolDissolveContext) CreateTx(mp minipool.IMinipool, opts *bind.TransactOpts) (types.ResponseStatus, *eth.TransactionInfo, error) {
	txInfo, err := mp.Common().Dissolve(opts)
	if err != nil {
		return types.ResponseStatus_Error, nil, err
	}
	return types.ResponseStatus_Success, txInfo, nil
}
