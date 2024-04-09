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
	"github.com/rocket-pool/rocketpool-go/v2/dao/proposals"
	"github.com/rocket-pool/rocketpool-go/v2/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/v2/dao/security"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/v2/types"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

const (
	propStateBatchSize int = 200
)

// ===============
// === Factory ===
// ===============

type securityStatusContextFactory struct {
	handler *SecurityCouncilHandler
}

func (f *securityStatusContextFactory) Create(args url.Values) (*securityStatusContext, error) {
	c := &securityStatusContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *securityStatusContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*securityStatusContext, api.SecurityStatusData](
		router, "status", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type securityStatusContext struct {
	handler     *SecurityCouncilHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	dpm      *proposals.DaoProposalManager
	pdaoMgr  *protocol.ProtocolDaoManager
	scMgr    *security.SecurityCouncilManager
	scMember *security.SecurityCouncilMember
}

func (c *securityStatusContext) Initialize() (types.ResponseStatus, error) {
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
	c.scMember, err = security.NewSecurityCouncilMember(c.rp, c.nodeAddress)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating security council member binding for %s: %w", c.nodeAddress.Hex(), err)
	}
	c.dpm, err = proposals.NewDaoProposalManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating DAO proposal manager binding: %w", err)
	}
	return types.ResponseStatus_Success, nil
}

func (c *securityStatusContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.scMember.Exists,
		c.scMember.InvitedTime,
		c.scMember.LeftTime,
		c.scMgr.MemberCount,
		c.dpm.ProposalCount,
		c.pdaoMgr.Settings.Security.ProposalActionTime,
	)
}

func (c *securityStatusContext) PrepareData(data *api.SecurityStatusData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	// Get member stats
	data.IsMember = c.scMember.Exists.Get()
	if data.IsMember {
		actionTime := c.pdaoMgr.Settings.Security.ProposalActionTime.Formatted()
		joinTime := c.scMember.InvitedTime.Formatted()
		leaveTime := c.scMember.LeftTime.Formatted()
		data.CanJoin = (time.Until(joinTime.Add(actionTime)) > 0)
		data.CanLeave = (time.Until(leaveTime.Add(actionTime)) > 0)
	}
	data.TotalMembers = c.scMgr.MemberCount.Formatted()

	// Get prop statuses
	propCount := c.dpm.ProposalCount.Formatted()
	_, props, err := c.dpm.GetProposals(propCount, false, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting proposals: %w", err)
	}
	err = c.rp.BatchQuery(int(propCount), propStateBatchSize, func(mc *batch.MultiCaller, i int) error {
		props[i].State.AddToQuery(mc)
		return nil
	}, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting proposal states: %w", err)
	}

	data.ProposalCounts.Total = int(propCount)
	for _, prop := range props {
		switch prop.State.Formatted() {
		case rptypes.ProposalState_Active:
			data.ProposalCounts.Active++
		case rptypes.ProposalState_Cancelled:
			data.ProposalCounts.Cancelled++
		case rptypes.ProposalState_Defeated:
			data.ProposalCounts.Defeated++
		case rptypes.ProposalState_Executed:
			data.ProposalCounts.Executed++
		case rptypes.ProposalState_Expired:
			data.ProposalCounts.Expired++
		case rptypes.ProposalState_Pending:
			data.ProposalCounts.Pending++
		case rptypes.ProposalState_Succeeded:
			data.ProposalCounts.Succeeded++
		}
	}
	return types.ResponseStatus_Success, nil
}
