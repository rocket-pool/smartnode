package security

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/dao/proposals"
	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/dao/security"
	"github.com/rocket-pool/rocketpool-go/rocketpool"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type securityLeaveContextFactory struct {
	handler *SecurityCouncilHandler
}

func (f *securityLeaveContextFactory) Create(vars map[string]string) (*securityLeaveContext, error) {
	c := &securityLeaveContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *securityLeaveContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*securityLeaveContext, api.SecurityLeaveData](
		router, "leave", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type securityLeaveContext struct {
	handler     *SecurityCouncilHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	scMgr     *security.SecurityCouncilManager
	scMember  *security.SecurityCouncilMember
	dpm       *proposals.DaoProposalManager
	pSettings *protocol.ProtocolDaoSettings
}

func (c *securityLeaveContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Bindings
	var err error
	c.scMember, err = security.NewSecurityCouncilMember(c.rp, c.nodeAddress)
	if err != nil {
		return fmt.Errorf("error creating security council member binding: %w", err)
	}
	c.dpm, err = proposals.NewDaoProposalManager(c.rp)
	if err != nil {
		return fmt.Errorf("error creating DAO proposal manager binding: %w", err)
	}
	pdaoMgr, err := protocol.NewProtocolDaoManager(c.rp)
	if err != nil {
		return fmt.Errorf("error creating protocol DAO manager binding: %w", err)
	}
	c.pSettings = pdaoMgr.Settings
	c.scMgr, err = security.NewSecurityCouncilManager(c.rp, c.pSettings)
	if err != nil {
		return fmt.Errorf("error creating security council manager binding: %w", err)
	}
	return nil
}

func (c *securityLeaveContext) GetState(mc *batch.MultiCaller) {
	core.AddQueryablesToMulticall(mc,
		c.scMember.Exists,
		c.scMember.LeftTime,
		c.pSettings.Security.ProposalActionTime,
	)
}

func (c *securityLeaveContext) PrepareData(data *api.SecurityLeaveData, opts *bind.TransactOpts) error {
	leftTime := c.scMember.LeftTime.Formatted()
	actionTime := c.pSettings.Security.ProposalActionTime.Formatted()
	data.ProposalExpired = time.Until(leftTime.Add(actionTime)) < 0
	data.IsNotMember = !c.scMember.Exists.Get()
	data.CanLeave = !(data.ProposalExpired || data.IsNotMember)

	// Get the tx
	if data.CanLeave && opts != nil {
		txInfo, err := c.scMgr.Leave(opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for Leave: %w", err)
		}
		data.TxInfo = txInfo
	}
	return nil
}
