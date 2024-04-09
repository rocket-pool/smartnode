package odao

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/dao/oracle"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type oracleDaoLeaveContextFactory struct {
	handler *OracleDaoHandler
}

func (f *oracleDaoLeaveContextFactory) Create(args url.Values) (*oracleDaoLeaveContext, error) {
	c := &oracleDaoLeaveContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("bondRefundAddress", args, input.ValidateAddress, &c.bondRefundAddress),
	}
	return c, errors.Join(inputErrs...)
}

func (f *oracleDaoLeaveContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*oracleDaoLeaveContext, api.OracleDaoLeaveData](
		router, "leave", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type oracleDaoLeaveContext struct {
	handler     *OracleDaoHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	bondRefundAddress common.Address
	odaoMember        *oracle.OracleDaoMember
	odaoMgr           *oracle.OracleDaoManager
	oSettings         *oracle.OracleDaoSettings
}

func (c *oracleDaoLeaveContext) Initialize() (types.ResponseStatus, error) {
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
	return types.ResponseStatus_Success, nil
}

func (c *oracleDaoLeaveContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.odaoMember.LeftTime,
		c.oSettings.Proposal.ActionTime,
		c.odaoMgr.MemberCount,
		c.odaoMgr.MinimumMemberCount,
	)
}

func (c *oracleDaoLeaveContext) PrepareData(data *api.OracleDaoLeaveData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	// Get the timestamp of the latest block
	ctx := c.handler.ctx
	latestHeader, err := c.rp.Client.HeaderByNumber(ctx, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting latest block header: %w", err)
	}
	currentTime := time.Unix(int64(latestHeader.Time), 0)
	actionWindow := c.oSettings.Proposal.ActionTime.Formatted()

	// Check proposal details
	membersCanLeave := (c.odaoMgr.MemberCount.Formatted() > c.odaoMgr.MinimumMemberCount.Formatted())
	data.InsufficientMembers = !membersCanLeave
	data.ProposalExpired = !isProposalActionable(actionWindow, c.odaoMember.InvitedTime.Formatted(), currentTime)
	data.CanLeave = !(data.ProposalExpired || data.InsufficientMembers)

	// Get the tx
	if data.CanLeave && opts != nil {
		txInfo, err := c.odaoMgr.Leave(c.bondRefundAddress, opts)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for Leave: %w", err)
		}
		data.TxInfo = txInfo
	}
	return types.ResponseStatus_Success, nil
}
