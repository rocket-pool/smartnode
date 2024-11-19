package wallet

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/validator"
	"net/url"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type walletRebuildContextFactory struct {
	handler *WalletHandler
}

func (f *walletRebuildContextFactory) Create(args url.Values) (*walletRebuildContext, error) {
	c := &walletRebuildContext{
		handler: f.handler,
	}
	inputError := server.ValidateOptionalArg("enable-partial-rebuild", args, input.ValidateBool, &c.enablePartialRebuild, nil)

	return c, inputError
}

func (f *walletRebuildContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*walletRebuildContext, api.WalletRebuildData](
		router, "rebuild", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type walletRebuildContext struct {
	handler              *WalletHandler
	enablePartialRebuild bool
}

func (c *walletRebuildContext) PrepareData(data *api.WalletRebuildData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	vMgr := sp.GetValidatorManager()
	keyRecoveryManager := validator.NewKeyRecoveryManager(vMgr, c.enablePartialRebuild, false)

	// Requirements
	err := sp.RequireWalletReady()
	if err != nil {
		return types.ResponseStatus_WalletNotReady, err
	}
	status, err := sp.RequireRocketPoolContracts(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Recover validator keys
	data.RebuiltValidatorKeys, data.FailureReasons, err = keyRecoveryManager.RecoverMinipoolKeys()
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error recovering minipool keys: %w", err)
	}
	return types.ResponseStatus_Success, nil
}
