package tx

import (
	"errors"
	"fmt"
	"net/url"
	_ "time/tzdata"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"

	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
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
		server.ValidateArg("hash", args, input.ValidateTxHash, &c.hash),
	}
	return c, errors.Join(inputErrs...)
}

func (f *txWaitContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*txWaitContext, api.SuccessData](
		router, "wait", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type txWaitContext struct {
	handler *TxHandler
	hash    common.Hash
}

func (c *txWaitContext) PrepareData(data *api.SuccessData, opts *bind.TransactOpts) error {
	sp := c.handler.serviceProvider
	rp := sp.GetRocketPool()

	err := rp.WaitForTransactionByHash(c.hash)
	if err != nil {
		return fmt.Errorf("error waiting for tx %s: %w", c.hash.Hex(), err)
	}

	data.Success = true
	return nil
}
