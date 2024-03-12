package minipool

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
)

// ===============
// === Factory ===
// ===============

type minipoolPromoteContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolPromoteContextFactory) Create(args url.Values) (*minipoolPromoteContext, error) {
	c := &minipoolPromoteContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArgBatch("addresses", args, minipoolAddressBatchSize, input.ValidateAddress, &c.minipoolAddresses),
	}
	return c, errors.Join(inputErrs...)
}

func (f *minipoolPromoteContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*minipoolPromoteContext, api.BatchTxInfoData](
		router, "promote", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolPromoteContext struct {
	handler           *MinipoolHandler
	minipoolAddresses []common.Address
}

func (c *minipoolPromoteContext) PrepareData(data *api.BatchTxInfoData, opts *bind.TransactOpts) error {
	return prepareMinipoolBatchTxData(c.handler.serviceProvider, c.minipoolAddresses, data, c.CreateTx, "promote")
}

func (c *minipoolPromoteContext) CreateTx(mp minipool.IMinipool, opts *bind.TransactOpts) (*eth.TransactionInfo, error) {
	mpv3, success := minipool.GetMinipoolAsV3(mp)
	if !success {
		mpCommon := mp.Common()
		return nil, fmt.Errorf("cannot create v3 binding for minipool %s, version %d", mpCommon.Address.Hex(), mpCommon.Version)
	}
	return mpv3.Promote(opts)
}
