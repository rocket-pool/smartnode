package minipool

import (
	"errors"
	"fmt"

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

type minipoolReduceBondContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolReduceBondContextFactory) Create(vars map[string]string) (*minipoolReduceBondContext, error) {
	c := &minipoolReduceBondContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("addresses", vars, cliutils.ValidateAddresses, &c.minipoolAddresses),
	}
	return c, errors.Join(inputErrs...)
}

func (f *minipoolReduceBondContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessRoute[*minipoolReduceBondContext, api.BatchTxInfoData](
		router, "reduce-bond", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolReduceBondContext struct {
	handler           *MinipoolHandler
	minipoolAddresses []common.Address
}

func (c *minipoolReduceBondContext) PrepareData(data *api.BatchTxInfoData, opts *bind.TransactOpts) error {
	return prepareMinipoolBatchTxData(c.handler.serviceProvider, c.minipoolAddresses, data, c.CreateTx, "reduce-bond")
}

func (c *minipoolReduceBondContext) CreateTx(mp minipool.Minipool, opts *bind.TransactOpts) (*core.TransactionInfo, error) {
	mpv3, success := minipool.GetMinipoolAsV3(mp)
	if !success {
		mpCommon := mp.GetMinipoolCommon()
		return nil, fmt.Errorf("cannot create v3 binding for minipool %s, version %d", mpCommon.Details.Address.Hex(), mpCommon.Details.Version)
	}
	return mpv3.ReduceBondAmount(opts)
}
