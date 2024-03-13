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

type minipoolReduceBondContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolReduceBondContextFactory) Create(args url.Values) (*minipoolReduceBondContext, error) {
	c := &minipoolReduceBondContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArgBatch("addresses", args, minipoolAddressBatchSize, input.ValidateAddress, &c.minipoolAddresses),
	}
	return c, errors.Join(inputErrs...)
}

func (f *minipoolReduceBondContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*minipoolReduceBondContext, types.BatchTxInfoData](
		router, "reduce-bond", f, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolReduceBondContext struct {
	handler           *MinipoolHandler
	minipoolAddresses []common.Address
}

func (c *minipoolReduceBondContext) PrepareData(data *types.BatchTxInfoData, opts *bind.TransactOpts) error {
	return prepareMinipoolBatchTxData(c.handler.context, c.handler.serviceProvider, c.minipoolAddresses, data, c.CreateTx, "reduce-bond")
}

func (c *minipoolReduceBondContext) CreateTx(mp minipool.IMinipool, opts *bind.TransactOpts) (*eth.TransactionInfo, error) {
	mpv3, success := minipool.GetMinipoolAsV3(mp)
	if !success {
		mpCommon := mp.Common()
		return nil, fmt.Errorf("cannot create v3 binding for minipool %s, version %d", mpCommon.Address.Hex(), mpCommon.Version)
	}
	return mpv3.ReduceBondAmount(opts)
}
