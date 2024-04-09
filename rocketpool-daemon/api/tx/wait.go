package tx

import (
	"errors"
	"fmt"
	"net/url"
	_ "time/tzdata"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/utils/input"
)

// ===============
// === Factory ===
// ===============

type txWaitContextFactory struct {
	handler *TxHandler
}

func (f *txWaitContextFactory) Create(args url.Values) (*txWaitContext, error) {
	c := &txWaitContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("hash", args, input.ValidateHash, &c.hash),
	}
	return c, errors.Join(inputErrs...)
}

func (f *txWaitContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*txWaitContext, types.SuccessData](
		router, "wait", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type txWaitContext struct {
	handler *TxHandler
	hash    common.Hash
}

func (c *txWaitContext) PrepareData(data *types.SuccessData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	rp := sp.GetRocketPool()

	err := rp.WaitForTransactionByHash(c.hash)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error waiting for tx %s: %w", c.hash.Hex(), err)
	}
	return types.ResponseStatus_Success, nil
}
