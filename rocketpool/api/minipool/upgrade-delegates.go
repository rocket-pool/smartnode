package minipool

import (
	"errors"

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

type minipoolUpgradeDelegatesContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolUpgradeDelegatesContextFactory) Create(vars map[string]string) (*minipoolUpgradeDelegatesContext, error) {
	c := &minipoolUpgradeDelegatesContext{
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

type minipoolUpgradeDelegatesContext struct {
	handler           *MinipoolHandler
	minipoolAddresses []common.Address
}

func (c *minipoolUpgradeDelegatesContext) PrepareData(data *api.BatchTxInfoData, opts *bind.TransactOpts) error {
	return prepareMinipoolBatchTxData(c.handler.serviceProvider, c.minipoolAddresses, data, c.CreateTx, "upgrade-delegate")
}

func (c *minipoolUpgradeDelegatesContext) CreateTx(mp minipool.Minipool, opts *bind.TransactOpts) (*core.TransactionInfo, error) {
	mpCommon := mp.GetMinipoolCommon()
	return mpCommon.DelegateUpgrade(opts)
}
