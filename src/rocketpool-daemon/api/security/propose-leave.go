package security

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/dao/security"

	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
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
	server.RegisterQuerylessGet[*securityProposeLeaveContext, api.TxInfoData](
		router, "propose-leave", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type securityProposeLeaveContext struct {
	handler *SecurityCouncilHandler
}

func (c *securityProposeLeaveContext) PrepareData(data *api.TxInfoData, opts *bind.TransactOpts) error {
	sp := c.handler.serviceProvider
	rp := sp.GetRocketPool()

	// Requirements
	err := sp.RequireOnSecurityCouncil()
	if err != nil {
		return err
	}

	// Bindings
	pdaoMgr, err := protocol.NewProtocolDaoManager(rp)
	if err != nil {
		return fmt.Errorf("error creating protocol DAO manager binding: %w", err)
	}
	pSettings := pdaoMgr.Settings
	scMgr, err := security.NewSecurityCouncilManager(rp, pSettings)
	if err != nil {
		return fmt.Errorf("error creating security council manager binding: %w", err)
	}

	// Get the tx
	if opts != nil {
		txInfo, err := scMgr.RequestLeave(opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for RequestLeave: %w", err)
		}
		data.TxInfo = txInfo
	}
	return nil
}
