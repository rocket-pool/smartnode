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
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type oracleDaoProposeLeaveContextFactory struct {
	handler *OracleDaoHandler
}

func (f *oracleDaoProposeLeaveContextFactory) Create(args url.Values) (*oracleDaoProposeLeaveContext, error) {
	c := &oracleDaoProposeLeaveContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *oracleDaoProposeLeaveContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*oracleDaoProposeLeaveContext, api.OracleDaoProposeLeaveData](
		router, "propose-leave", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type oracleDaoProposeLeaveContext struct {
	handler     *OracleDaoHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	odaoMember *oracle.OracleDaoMember
	oSettings  *oracle.OracleDaoSettings
	odaoMgr    *oracle.OracleDaoManager
}

func (c *oracleDaoProposeLeaveContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	status, err := sp.RequireOnOracleDao(c.handler.ctx)
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

func (c *oracleDaoProposeLeaveContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.odaoMember.LastProposalTime,
		c.odaoMember.ID,
		c.odaoMember.Url,
		c.odaoMgr.MemberCount,
		c.odaoMgr.MinimumMemberCount,
	)
	c.oSettings.Proposal.CooldownTime.AddToQuery(mc)
}

func (c *oracleDaoProposeLeaveContext) PrepareData(data *api.OracleDaoProposeLeaveData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	// Get the timestamp of the latest block
	ctx := c.handler.ctx
	latestHeader, err := c.rp.Client.HeaderByNumber(ctx, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting latest block header: %w", err)
	}
	currentTime := time.Unix(int64(latestHeader.Time), 0)
	cooldownTime := c.oSettings.Proposal.CooldownTime.Formatted()

	// Check proposal details
	data.ProposalCooldownActive = isProposalCooldownActive(cooldownTime, c.odaoMember.LastProposalTime.Formatted(), currentTime)
	data.InsufficientMembers = c.odaoMgr.MemberCount.Formatted() <= c.odaoMgr.MinimumMemberCount.Formatted()
	data.CanPropose = !(data.ProposalCooldownActive || data.InsufficientMembers)

	// Get the tx
	if data.CanPropose && opts != nil {
		message := fmt.Sprintf("%s (%s) leaves", c.odaoMember.ID.Get(), c.odaoMember.Url.Get())
		txInfo, err := c.odaoMgr.ProposeMemberLeave(message, c.nodeAddress, opts)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for ProposeMemberLeave: %w", err)
		}
		data.TxInfo = txInfo
	}
	return types.ResponseStatus_Success, nil
}
