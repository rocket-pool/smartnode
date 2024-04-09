package security

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/rocketpool-go/v2/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/v2/dao/security"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
)

// ===============
// === Factory ===
// ===============

type securityProposeLeaveContextFactory struct {
	handler *SecurityCouncilHandler
}

func (f *securityProposeLeaveContextFactory) Create(args url.Values) (*securityProposeLeaveContext, error) {
	c := &securityProposeLeaveContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *securityProposeLeaveContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*securityProposeLeaveContext, types.TxInfoData](
		router, "propose-leave", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type securityProposeLeaveContext struct {
	handler *SecurityCouncilHandler
}

func (c *securityProposeLeaveContext) PrepareData(data *types.TxInfoData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	rp := sp.GetRocketPool()

	// Requirements
	status, err := sp.RequireOnSecurityCouncil(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	pdaoMgr, err := protocol.NewProtocolDaoManager(rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating protocol DAO manager binding: %w", err)
	}
	pSettings := pdaoMgr.Settings
	scMgr, err := security.NewSecurityCouncilManager(rp, pSettings)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating security council manager binding: %w", err)
	}

	// Get the tx
	if opts != nil {
		txInfo, err := scMgr.RequestLeave(opts)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for RequestLeave: %w", err)
		}
		data.TxInfo = txInfo
	}
	return types.ResponseStatus_Success, nil
}
