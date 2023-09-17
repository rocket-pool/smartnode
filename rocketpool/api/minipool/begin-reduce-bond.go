package minipool

import (
	"errors"
	"fmt"
	"math/big"

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

type minipoolBeginReduceBondContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolBeginReduceBondContextFactory) Create(vars map[string]string) (*minipoolBeginReduceBondContext, error) {
	c := &minipoolBeginReduceBondContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("newBondAmount", vars, cliutils.ValidateBigInt, &c.newBondAmountWei),
		server.ValidateArg("addresses", vars, cliutils.ValidateAddresses, &c.minipoolAddresses),
	}
	return c, errors.Join(inputErrs...)
}

func (f *minipoolBeginReduceBondContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessRoute[*minipoolBeginReduceBondContext, api.BatchTxInfoData](
		router, "begin-reduce-bond", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolBeginReduceBondContext struct {
	handler           *MinipoolHandler
	newBondAmountWei  *big.Int
	minipoolAddresses []common.Address
}

func (c *minipoolBeginReduceBondContext) PrepareData(data *api.BatchTxInfoData, opts *bind.TransactOpts) error {
	return prepareMinipoolBatchTxData(c.handler.serviceProvider, c.minipoolAddresses, data, c.CreateTx, "begin-bond-reduce")
}

func (c *minipoolBeginReduceBondContext) CreateTx(mp minipool.Minipool, opts *bind.TransactOpts) (*core.TransactionInfo, error) {
	mpv3, success := minipool.GetMinipoolAsV3(mp)
	if !success {
		mpCommon := mp.GetMinipoolCommon()
		return nil, fmt.Errorf("cannot create v3 binding for minipool %s, version %d", mpCommon.Details.Address.Hex(), mpCommon.Details.Version)
	}
	return mpv3.BeginReduceBondAmount(c.newBondAmountWei, opts)
}
