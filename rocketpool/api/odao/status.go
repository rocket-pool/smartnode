package odao

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/dao"
	"github.com/rocket-pool/rocketpool-go/dao/oracle"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/types"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type oracleDaoStatusContextFactory struct {
	handler *OracleDaoHandler
}

func (f *oracleDaoStatusContextFactory) Create(vars map[string]string) (*oracleDaoStatusContext, error) {
	c := &oracleDaoStatusContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *oracleDaoStatusContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*oracleDaoStatusContext, api.OracleDaoStatusData](
		router, "status", f, f.handler.serviceProvider,
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
	dpm        *dao.DaoProposalManager
}

func (c *oracleDaoStatusContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	err := sp.RequireNodeRegistered()
	if err != nil {
		return err
	}

	// Bindings
	c.odaoMember, err = oracle.NewOracleDaoMember(c.rp, c.nodeAddress)
	if err != nil {
		return fmt.Errorf("error creating oracle DAO member binding: %w", err)
	}
	c.odaoMgr, err = oracle.NewOracleDaoManager(c.rp)
	if err != nil {
		return fmt.Errorf("error creating Oracle DAO manager binding: %w", err)
	}
	c.oSettings = c.odaoMgr.Settings
	c.dpm, err = dao.NewDaoProposalManager(c.rp)
	if err != nil {
		return fmt.Errorf("error creating proposal manager binding: %w", err)
	}
	return nil
}

func (c *oracleDaoStatusContext) GetState(mc *batch.MultiCaller) {
	c.odaoMember.GetExists(mc)
	c.odaoMember.GetInvitedTime(mc)
	c.odaoMember.GetReplacedTime(mc)
	c.odaoMember.GetLeftTime(mc)
	c.odaoMgr.GetMemberCount(mc)
	c.dpm.GetProposalCount(mc)
	c.oSettings.GetProposalActionTime(mc)
}

func (c *oracleDaoStatusContext) PrepareData(data *api.OracleDaoStatusData, opts *bind.TransactOpts) error {
	// Get the timestamp of the latest block
	latestHeader, err := c.rp.Client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("error getting latest block header: %w", err)
	}
	currentTime := time.Unix(int64(latestHeader.Time), 0)
	actionWindow := c.oSettings.Details.Proposals.ActionTime.Formatted()

	// Check action windows for the current member
	exists := c.odaoMember.Details.Exists
	data.IsMember = exists
	if exists {
		data.CanLeave = isProposalActionable(actionWindow, c.odaoMember.Details.LeftTime.Formatted(), currentTime)
		data.CanReplace = isProposalActionable(actionWindow, c.odaoMember.Details.ReplacedTime.Formatted(), currentTime)
	} else {
		data.CanJoin = isProposalActionable(actionWindow, c.odaoMember.Details.InvitedTime.Formatted(), currentTime)
	}

	// Total member count
	data.TotalMembers = c.odaoMgr.Details.MemberCount.Formatted()

	// Get the proposals
	_, props, err := c.dpm.GetProposals(c.rp, c.dpm.Details.ProposalCount.Formatted(), false, nil)
	if err != nil {
		return fmt.Errorf("error getting Oracle DAO proposals: %w", err)
	}

	// Proposal info
	data.ProposalCounts.Total = len(props)
	for _, prop := range props {
		switch prop.Details.State.Formatted() {
		case rptypes.Pending:
			data.ProposalCounts.Pending++
		case rptypes.Active:
			data.ProposalCounts.Active++
		case rptypes.Cancelled:
			data.ProposalCounts.Cancelled++
		case rptypes.Defeated:
			data.ProposalCounts.Defeated++
		case rptypes.Succeeded:
			data.ProposalCounts.Succeeded++
		case rptypes.Expired:
			data.ProposalCounts.Expired++
		case rptypes.Executed:
			data.ProposalCounts.Executed++
		}
	}
	return nil
}
