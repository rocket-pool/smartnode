package network

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/utils/input"
	rputils "github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/utils"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type networkDepositInfoContextFactory struct {
	handler *NetworkHandler
}

func (f *networkDepositInfoContextFactory) Create(args url.Values) (*networkDepositInfoContext, error) {
	c := &networkDepositInfoContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("checkSync", args, input.ValidateBool, &c.checkSync),
	}
	return c, errors.Join(inputErrs...)
}

func (f *networkDepositInfoContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*networkDepositInfoContext, api.NetworkDepositContractInfoData](
		router, "deposit-contract-info", f, f.handler.logger.Logger, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type networkDepositInfoContext struct {
	handler   *NetworkHandler
	checkSync bool
}

func (c *networkDepositInfoContext) PrepareData(data *api.NetworkDepositContractInfoData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	rp := sp.GetRocketPool()
	cfg := sp.GetConfig()
	res := sp.GetResources()
	bc := sp.GetBeaconClient()
	ctx := c.handler.ctx

	// Requirements
	if c.checkSync {
		// Check the EL sync progress and the RP contract cache
		status, err := sp.RequireRocketPoolContracts(c.handler.ctx)
		if err != nil {
			return status, err
		}
	} else {
		// Just check the RP contract cache
		err := sp.RefreshRocketPoolContracts()
		if err != nil {
			return types.ResponseStatus_Error, err
		}
	}

	// Get the deposit contract info
	info, err := rputils.GetDepositContractInfo(ctx, rp, cfg, res, bc)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting deposit contract info: %w", err)
	}
	data.SufficientSync = true
	data.RPNetwork = info.RPNetwork
	data.RPDepositContract = info.RPDepositContract
	data.BeaconNetwork = info.BeaconNetwork
	data.BeaconDepositContract = info.BeaconDepositContract
	return types.ResponseStatus_Success, nil
}
