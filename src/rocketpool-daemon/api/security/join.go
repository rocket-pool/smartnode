package security

import (
	"fmt"
	"net/url"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/v2/dao/security"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type securityJoinContextFactory struct {
	handler *SecurityCouncilHandler
}

func (f *securityJoinContextFactory) Create(args url.Values) (*securityJoinContext, error) {
	c := &securityJoinContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *securityJoinContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*securityJoinContext, api.SecurityJoinData](
		router, "join", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type securityJoinContext struct {
	handler     *SecurityCouncilHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	scMgr     *security.SecurityCouncilManager
	scMember  *security.SecurityCouncilMember
	pSettings *protocol.ProtocolDaoSettings
}

func (c *securityJoinContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	status, err := sp.RequireNodeRegistered(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	c.scMember, err = security.NewSecurityCouncilMember(c.rp, c.nodeAddress)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating security council member binding: %w", err)
	}
	pdaoMgr, err := protocol.NewProtocolDaoManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating protocol DAO manager binding: %w", err)
	}
	c.pSettings = pdaoMgr.Settings
	c.scMgr, err = security.NewSecurityCouncilManager(c.rp, c.pSettings)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating security council manager binding: %w", err)
	}
	return types.ResponseStatus_Success, nil
}

func (c *securityJoinContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.scMember.Exists,
		c.scMember.InvitedTime,
		c.pSettings.Security.ProposalActionTime,
	)
}

func (c *securityJoinContext) PrepareData(data *api.SecurityJoinData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	invitedTime := c.scMember.InvitedTime.Formatted()
	actionTime := c.pSettings.Security.ProposalActionTime.Formatted()
	data.ProposalExpired = time.Until(invitedTime.Add(actionTime)) < 0
	data.AlreadyMember = c.scMember.Exists.Get()
	data.CanJoin = !(data.ProposalExpired || data.AlreadyMember)

	// Get the tx
	if data.CanJoin && opts != nil {
		txInfo, err := c.scMgr.Join(opts)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for Join: %w", err)
		}
		data.TxInfo = txInfo
	}
	return types.ResponseStatus_Success, nil
}
