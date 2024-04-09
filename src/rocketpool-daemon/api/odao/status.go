package odao

import (
	"fmt"
	"net/url"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/dao/oracle"
	"github.com/rocket-pool/rocketpool-go/v2/dao/proposals"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/v2/types"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type oracleDaoStatusContextFactory struct {
	handler *OracleDaoHandler
}

func (f *oracleDaoStatusContextFactory) Create(args url.Values) (*oracleDaoStatusContext, error) {
	c := &oracleDaoStatusContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *oracleDaoStatusContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*oracleDaoStatusContext, api.OracleDaoStatusData](
		router, "status", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type oracleDaoStatusContext struct {
	handler     *OracleDaoHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	odaoMember *oracle.OracleDaoMember
	oSettings  *oracle.OracleDaoSettings
	odaoMgr    *oracle.OracleDaoManager
	dpm        *proposals.DaoProposalManager
}

func (c *oracleDaoStatusContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	status, err := sp.RequireNodeRegistered(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	c.odaoMember, err = oracle.NewOracleDaoMember(c.rp, c.nodeAddress)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating oracle DAO member binding: %w", err)
	}
	c.odaoMgr, err = oracle.NewOracleDaoManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating Oracle DAO manager binding: %w", err)
	}
	c.oSettings = c.odaoMgr.Settings
	c.dpm, err = proposals.NewDaoProposalManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating proposal manager binding: %w", err)
	}
	return types.ResponseStatus_Success, nil
}

func (c *oracleDaoStatusContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.odaoMember.Exists,
		c.odaoMember.InvitedTime,
		c.odaoMember.ReplacedTime,
		c.odaoMember.LeftTime,
		c.odaoMgr.MemberCount,
		c.dpm.ProposalCount,
	)
	c.oSettings.Proposal.ActionTime.AddToQuery(mc)
}

func (c *oracleDaoStatusContext) PrepareData(data *api.OracleDaoStatusData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	// Get the timestamp of the latest block
	ctx := c.handler.ctx
	latestHeader, err := c.rp.Client.HeaderByNumber(ctx, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting latest block header: %w", err)
	}
	currentTime := time.Unix(int64(latestHeader.Time), 0)
	actionWindow := c.oSettings.Proposal.ActionTime.Formatted()

	// Check action windows for the current member
	exists := c.odaoMember.Exists.Get()
	data.IsMember = exists
	if exists {
		data.CanLeave = isProposalActionable(actionWindow, c.odaoMember.LeftTime.Formatted(), currentTime)
		data.CanReplace = isProposalActionable(actionWindow, c.odaoMember.ReplacedTime.Formatted(), currentTime)
	} else {
		data.CanJoin = isProposalActionable(actionWindow, c.odaoMember.InvitedTime.Formatted(), currentTime)
	}

	// Total member count
	data.TotalMembers = c.odaoMgr.MemberCount.Formatted()

	// Get the proposals
	_, props, err := c.dpm.GetProposals(c.dpm.ProposalCount.Formatted(), false, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting Oracle DAO proposals: %w", err)
	}

	// Proposal info
	data.ProposalCounts.Total = len(props)
	for _, prop := range props {
		switch prop.State.Formatted() {
		case rptypes.ProposalState_Pending:
			data.ProposalCounts.Pending++
		case rptypes.ProposalState_Active:
			data.ProposalCounts.Active++
		case rptypes.ProposalState_Cancelled:
			data.ProposalCounts.Cancelled++
		case rptypes.ProposalState_Defeated:
			data.ProposalCounts.Defeated++
		case rptypes.ProposalState_Succeeded:
			data.ProposalCounts.Succeeded++
		case rptypes.ProposalState_Expired:
			data.ProposalCounts.Expired++
		case rptypes.ProposalState_Executed:
			data.ProposalCounts.Executed++
		}
	}
	return types.ResponseStatus_Success, nil
}
