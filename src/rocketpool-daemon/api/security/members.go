package security

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/node-manager-core/eth"

	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/rocketpool-go/v2/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/v2/dao/security"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type securityMembersContextFactory struct {
	handler *SecurityCouncilHandler
}

func (f *securityMembersContextFactory) Create(args url.Values) (*securityMembersContext, error) {
	c := &securityMembersContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *securityMembersContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*securityMembersContext, api.SecurityMembersData](
		router, "members", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type securityMembersContext struct {
	handler     *SecurityCouncilHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	scMgr *security.SecurityCouncilManager
}

func (c *securityMembersContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	status, err := sp.RequireRocketPoolContracts(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	pdaoMgr, err := protocol.NewProtocolDaoManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating protocol DAO manager binding: %w", err)
	}
	c.scMgr, err = security.NewSecurityCouncilManager(c.rp, pdaoMgr.Settings)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating security council manager binding: %w", err)
	}
	return types.ResponseStatus_Success, nil
}

func (c *securityMembersContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.scMgr.MemberCount,
	)
}

func (c *securityMembersContext) PrepareData(data *api.SecurityMembersData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	// Get all members
	memberCount := c.scMgr.MemberCount.Formatted()
	addresses, err := c.scMgr.GetMemberAddresses(memberCount, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting member addresses: %w", err)
	}
	members, err := c.scMgr.CreateMembersFromAddresses(addresses, true, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting member details: %w", err)
	}

	data.Members = make([]api.SecurityMemberDetails, memberCount)
	for i, details := range members {
		member := api.SecurityMemberDetails{
			Address:     details.Address,
			Exists:      details.Exists.Get(),
			ID:          details.ID.Get(),
			InvitedTime: details.InvitedTime.Formatted(),
			JoinedTime:  details.JoinedTime.Formatted(),
			LeftTime:    details.LeftTime.Formatted(),
		}
		data.Members[i] = member
	}
	return types.ResponseStatus_Success, nil
}
