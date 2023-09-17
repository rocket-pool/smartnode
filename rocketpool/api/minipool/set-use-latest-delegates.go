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

type minipoolSetUseLatestDelegatesContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolSetUseLatestDelegatesContextFactory) Create(vars map[string]string) (*minipoolSetUseLatestDelegatesContext, error) {
	c := &minipoolSetUseLatestDelegatesContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("addresses", vars, cliutils.ValidateAddresses, &c.minipoolAddresses),
		server.ValidateArg("setting", vars, cliutils.ValidateBool, &c.setting),
	}
	return c, errors.Join(inputErrs...)
}

func (f *minipoolSetUseLatestDelegatesContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessRoute[*minipoolSetUseLatestDelegatesContext, api.BatchTxInfoData](
		router, "delegate/set-use-latest", f, f.handler.serviceProvider,
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

func (c *minipoolSetUseLatestDelegatesContext) PrepareData(data *api.BatchTxInfoData, opts *bind.TransactOpts) error {
	return prepareMinipoolBatchTxData(c.handler.serviceProvider, c.minipoolAddresses, data, c.CreateTx, "set-use-latest-delegate")
}

func (c *minipoolSetUseLatestDelegatesContext) CreateTx(mp minipool.Minipool, opts *bind.TransactOpts) (*core.TransactionInfo, error) {
	mpCommon := mp.GetMinipoolCommon()
	return mpCommon.SetUseLatestDelegate(c.setting, opts)
}
