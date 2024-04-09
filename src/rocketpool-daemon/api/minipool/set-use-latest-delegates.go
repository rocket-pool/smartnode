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

type minipoolSetUseLatestDelegatesContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolSetUseLatestDelegatesContextFactory) Create(args url.Values) (*minipoolSetUseLatestDelegatesContext, error) {
	c := &minipoolSetUseLatestDelegatesContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArgBatch("addresses", args, minipoolAddressBatchSize, input.ValidateAddress, &c.minipoolAddresses),
		server.ValidateArg("setting", args, input.ValidateBool, &c.setting),
	}
	return c, errors.Join(inputErrs...)
}

func (f *minipoolSetUseLatestDelegatesContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*minipoolSetUseLatestDelegatesContext, types.BatchTxInfoData](
		router, "delegate/set-use-latest", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolSetUseLatestDelegatesContext struct {
	handler           *MinipoolHandler
	setting           bool
	minipoolAddresses []common.Address
}

func (c *minipoolSetUseLatestDelegatesContext) PrepareData(data *types.BatchTxInfoData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	return prepareMinipoolBatchTxData(c.handler.ctx, c.handler.serviceProvider, c.minipoolAddresses, data, c.CreateTx, "set-use-latest-delegate")
}

func (c *minipoolSetUseLatestDelegatesContext) CreateTx(mp minipool.IMinipool, opts *bind.TransactOpts) (types.ResponseStatus, *eth.TransactionInfo, error) {
	txInfo, err := mp.Common().SetUseLatestDelegate(c.setting, opts)
	if err != nil {
		return types.ResponseStatus_Error, nil, err
	}
	return types.ResponseStatus_Success, txInfo, nil
}
