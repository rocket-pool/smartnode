package minipool

import (
	"errors"
	"fmt"
	"math/big"
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

type minipoolBeginReduceBondContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolBeginReduceBondContextFactory) Create(args url.Values) (*minipoolBeginReduceBondContext, error) {
	c := &minipoolBeginReduceBondContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("new-bond-amount", args, input.ValidateBigInt, &c.newBondAmountWei),
		server.ValidateArgBatch("addresses", args, minipoolAddressBatchSize, input.ValidateAddress, &c.minipoolAddresses),
	}
	return c, errors.Join(inputErrs...)
}

func (f *minipoolBeginReduceBondContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*minipoolBeginReduceBondContext, types.BatchTxInfoData](
		router, "begin-reduce-bond", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
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

func (c *minipoolBeginReduceBondContext) PrepareData(data *types.BatchTxInfoData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	return prepareMinipoolBatchTxData(c.handler.ctx, c.handler.serviceProvider, c.minipoolAddresses, data, c.CreateTx, "begin-bond-reduce")
}

func (c *minipoolBeginReduceBondContext) CreateTx(mp minipool.IMinipool, opts *bind.TransactOpts) (types.ResponseStatus, *eth.TransactionInfo, error) {
	mpv3, success := minipool.GetMinipoolAsV3(mp)
	if !success {
		mpCommon := mp.Common()
		return types.ResponseStatus_InvalidChainState, nil, fmt.Errorf("cannot create v3 binding for minipool %s, version %d", mpCommon.Address.Hex(), mpCommon.Version)
	}

	txInfo, err := mpv3.BeginReduceBondAmount(c.newBondAmountWei, opts)
	if err != nil {
		return types.ResponseStatus_Error, nil, err
	}
	return types.ResponseStatus_Success, txInfo, nil
}
