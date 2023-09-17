package minipool

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// ===============
// === Factory ===
// ===============

type minipoolDistributeContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolDistributeContextFactory) Create(vars map[string]string) (*minipoolDistributeContext, error) {
	c := &minipoolDistributeContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("addresses", vars, cliutils.ValidateAddresses, &c.minipoolAddresses),
	}
	return c, errors.Join(inputErrs...)
}

// ===============
// === Context ===
// ===============

type minipoolDistributeContext struct {
	handler           *MinipoolHandler
	minipoolAddresses []common.Address
}

func (c *minipoolDistributeContext) PrepareData(data *api.BatchTxInfoData, opts *bind.TransactOpts) error {
	return prepareMinipoolBatchTxData(c.handler.serviceProvider, c.minipoolAddresses, data, c.CreateTx, "distribute-balance")
}

func (c *minipoolDistributeContext) CreateTx(mp minipool.Minipool, opts *bind.TransactOpts) (*core.TransactionInfo, error) {
	mpv3, success := minipool.GetMinipoolAsV3(mp)
	if !success {
		mpCommon := mp.GetMinipoolCommon()
		return nil, fmt.Errorf("cannot create v3 binding for minipool %s, version %d", mpCommon.Details.Address.Hex(), mpCommon.Details.Version)
	}
	return mpv3.DistributeBalance(opts, true)
}
