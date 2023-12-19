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

type minipoolUpgradeDelegatesContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolUpgradeDelegatesContextFactory) Create(args url.Values) (*minipoolUpgradeDelegatesContext, error) {
	c := &minipoolUpgradeDelegatesContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("addresses", args, input.ValidateAddresses, &c.minipoolAddresses),
	}
	return c, errors.Join(inputErrs...)
}

func (f *minipoolUpgradeDelegatesContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*minipoolUpgradeDelegatesContext, api.BatchTxInfoData](
		router, "delegate/upgrade", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolUpgradeDelegatesContext struct {
	handler           *MinipoolHandler
	minipoolAddresses []common.Address
}

func (c *minipoolUpgradeDelegatesContext) PrepareData(data *api.BatchTxInfoData, opts *bind.TransactOpts) error {
	return prepareMinipoolBatchTxData(c.handler.serviceProvider, c.minipoolAddresses, data, c.CreateTx, "upgrade-delegate")
}

func (c *minipoolUpgradeDelegatesContext) CreateTx(mp minipool.IMinipool, opts *bind.TransactOpts) (*core.TransactionInfo, error) {
	return mp.Common().DelegateUpgrade(opts)
}
