package odao

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/dao/oracle"
	"github.com/rocket-pool/rocketpool-go/rocketpool"

	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
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
		router, "propose-leave", f, f.handler.serviceProvider,
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

func (c *oracleDaoProposeLeaveContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	err := sp.RequireOnOracleDao()
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
	return nil
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

func (c *oracleDaoProposeLeaveContext) PrepareData(data *api.OracleDaoProposeLeaveData, opts *bind.TransactOpts) error {
	// Get the timestamp of the latest block
	latestHeader, err := c.rp.Client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("error getting latest block header: %w", err)
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
			return fmt.Errorf("error getting TX info for ProposeMemberLeave: %w", err)
		}
		data.TxInfo = txInfo
	}
	return nil
}
