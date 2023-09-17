package minipool

import (
	"errors"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// ===============
// === Factory ===
// ===============

type minipoolDissolveContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolDissolveContextFactory) Create(vars map[string]string) (*minipoolDissolveContext, error) {
	c := &minipoolDissolveContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("addresses", vars, cliutils.ValidateAddresses, &c.minipoolAddresses),
	}
	return c, errors.Join(inputErrs...)
}

func (f *minipoolDissolveContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessRoute[*minipoolDissolveContext, api.BatchTxInfoData](
		router, "dissolve", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolDissolveContext struct {
	handler           *MinipoolHandler
	minipoolAddresses []common.Address
}

func (c *minipoolDissolveContext) PrepareData(data *api.BatchTxInfoData, opts *bind.TransactOpts) error {
	return prepareMinipoolBatchTxData(c.handler.serviceProvider, c.minipoolAddresses, data, c.CreateTx, "dissolve")
}

func (c *minipoolDissolveContext) CreateTx(mp minipool.Minipool, opts *bind.TransactOpts) (*core.TransactionInfo, error) {
	mpCommon := mp.GetMinipoolCommon()
	return mpCommon.Dissolve(opts)
}
