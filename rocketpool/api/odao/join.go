package odao

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/dao/oracle"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/tokens"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type oracleDaoJoinContextFactory struct {
	handler *OracleDaoHandler
}

func (f *oracleDaoJoinContextFactory) Create(vars map[string]string) (*oracleDaoJoinContext, error) {
	c := &oracleDaoJoinContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *oracleDaoJoinContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*oracleDaoJoinContext, api.OracleDaoJoinData](
		router, "join", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type oracleDaoJoinContext struct {
	handler     *OracleDaoHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	odaoMember *oracle.OracleDaoMember
	odaoMgr    *oracle.OracleDaoManager
	oSettings  *oracle.OracleDaoSettings
	rpl        *tokens.TokenRpl
	rplBalance *big.Int
}

func (c *oracleDaoJoinContext) Initialize() error {
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
	c.rpl, err = tokens.NewTokenRpl(c.rp)
	if err != nil {
		return fmt.Errorf("error creating RPL token binding: %w", err)
	}
	c.odaoMgr, err = oracle.NewOracleDaoManager(c.rp)
	if err != nil {
		return fmt.Errorf("error creating Oracle DAO manager binding: %w", err)
	}
	c.oSettings = c.odaoMgr.Settings
	return nil
}

func (c *oracleDaoJoinContext) GetState(mc *batch.MultiCaller) {
	c.odaoMember.GetInvitedTime(mc)
	c.oSettings.Proposal.ActionTime.Get(mc)
	c.odaoMember.GetExists(mc)
	c.rpl.GetBalance(mc, &c.rplBalance, c.nodeAddress)
	c.oSettings.Member.RplBond.Get(mc)
}

func (c *oracleDaoJoinContext) PrepareData(data *api.OracleDaoJoinData, opts *bind.TransactOpts) error {
	// Get the timestamp of the latest block
	latestHeader, err := c.rp.Client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("error getting latest block header: %w", err)
	}
	currentTime := time.Unix(int64(latestHeader.Time), 0)
	actionWindow := c.oSettings.Proposal.ActionTime.Value.Formatted()
	rplBond := c.oSettings.Member.RplBond.Value

	// Check proposal details
	data.ProposalExpired = !isProposalActionable(actionWindow, c.odaoMember.InvitedTime.Formatted(), currentTime)
	data.AlreadyMember = c.odaoMember.Exists
	data.InsufficientRplBalance = (c.rplBalance.Cmp(rplBond) < 0)
	data.CanJoin = !(data.ProposalExpired || data.AlreadyMember || data.InsufficientRplBalance)

	// Get the tx
	if data.CanJoin && opts != nil {
		dnta, err := c.rp.GetContract(rocketpool.ContractName_RocketDAONodeTrustedActions)
		if err != nil {
			return fmt.Errorf("error getting RPL token contract: %w", err)
		}

		approveTxInfo, err := c.rpl.Approve(*dnta.Address, rplBond, opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for RPL approval: %w", err)
		}
		data.ApproveTxInfo = approveTxInfo

		joinTxInfo, err := c.odaoMgr.Join(opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for Join: %w", err)
		}
		data.JoinTxInfo = joinTxInfo
	}
	return nil
}
