package minipool

import (
	"errors"
	"math/big"

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

// ===============
// === Context ===
// ===============

type minipoolBeginReduceBondContext struct {
	handler           *MinipoolHandler
	newBondAmountWei  *big.Int
	minipoolAddresses []common.Address
}

func (c *minipoolBeginReduceBondContext) PrepareData(data *api.BatchTxInfoData, opts *bind.TransactOpts) error {
	createBatchTxResponseForV3(c.handler.serviceProvider, c.minipoolAddresses, data, c.CreateTx, "begin-bond-reduce")
	return nil
}

func (c *minipoolBeginReduceBondContext) CreateTx(mpv3 *minipool.MinipoolV3, opts *bind.TransactOpts) (*core.TransactionInfo, error) {
	return mpv3.BeginReduceBondAmount(c.newBondAmountWei, opts)
}
