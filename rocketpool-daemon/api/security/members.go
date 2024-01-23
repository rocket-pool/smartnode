package security

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"

	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/dao/security"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
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
		router, "members", f, f.handler.serviceProvider,
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

func (c *securityMembersContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Bindings
	pdaoMgr, err := protocol.NewProtocolDaoManager(c.rp)
	if err != nil {
		return fmt.Errorf("error creating protocol DAO manager binding: %w", err)
	}
	c.scMgr, err = security.NewSecurityCouncilManager(c.rp, pdaoMgr.Settings)
	if err != nil {
		return fmt.Errorf("error creating security council manager binding: %w", err)
	}
	return nil
}

func (c *securityMembersContext) GetState(mc *batch.MultiCaller) {
	core.AddQueryablesToMulticall(mc,
		c.scMgr.MemberCount,
	)
}

func (c *securityMembersContext) PrepareData(data *api.SecurityMembersData, opts *bind.TransactOpts) error {
	// Get all members
	memberCount := c.scMgr.MemberCount.Formatted()
	addresses, err := c.scMgr.GetMemberAddresses(memberCount, nil)
	if err != nil {
		return fmt.Errorf("error getting member addresses: %w", err)
	}
	members, err := c.scMgr.CreateMembersFromAddresses(addresses, true, nil)
	if err != nil {
		return fmt.Errorf("error getting member details: %w", err)
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
	return nil
}
