package minipool

import (
	"errors"
	"fmt"
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

type minipoolDistributeContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolDistributeContextFactory) Create(args url.Values) (*minipoolDistributeContext, error) {
	c := &minipoolDistributeContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArgBatch("addresses", args, minipoolAddressBatchSize, input.ValidateAddress, &c.minipoolAddresses),
	}
	return c, errors.Join(inputErrs...)
}

func (f *minipoolDistributeContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*minipoolDistributeContext, types.BatchTxInfoData](
		router, "distribute", f, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolDistributeContext struct {
	handler           *MinipoolHandler
	minipoolAddresses []common.Address
}

func (c *minipoolDistributeContext) PrepareData(data *types.BatchTxInfoData, opts *bind.TransactOpts) error {
	return prepareMinipoolBatchTxData(c.handler.serviceProvider, c.minipoolAddresses, data, c.CreateTx, "distribute-balance")
}

func (c *minipoolDistributeContext) CreateTx(mp minipool.IMinipool, opts *bind.TransactOpts) (*eth.TransactionInfo, error) {
	mpv3, success := minipool.GetMinipoolAsV3(mp)
	if !success {
		mpCommon := mp.Common()
		return nil, fmt.Errorf("cannot create v3 binding for minipool %s, version %d", mpCommon.Address.Hex(), mpCommon.Version)
	}
	return mpv3.DistributeBalance(opts, true)
}
