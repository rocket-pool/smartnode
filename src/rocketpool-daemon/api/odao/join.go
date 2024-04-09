package odao

import (
	"fmt"
	"math/big"
	"net/url"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/core"
	"github.com/rocket-pool/rocketpool-go/v2/dao/oracle"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/rocketpool-go/v2/tokens"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type oracleDaoJoinContextFactory struct {
	handler *OracleDaoHandler
}

func (f *oracleDaoJoinContextFactory) Create(args url.Values) (*oracleDaoJoinContext, error) {
	c := &oracleDaoJoinContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *oracleDaoJoinContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*oracleDaoJoinContext, api.OracleDaoJoinData](
		router, "join", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type oracleDaoJoinContext struct {
	handler     *OracleDaoHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address
	dnta        *core.Contract

	odaoMember   *oracle.OracleDaoMember
	odaoMgr      *oracle.OracleDaoManager
	oSettings    *oracle.OracleDaoSettings
	rpl          *tokens.TokenRpl
	rplBalance   *big.Int
	rplAllowance *big.Int
}

func (c *oracleDaoJoinContext) Initialize() (types.ResponseStatus, error) {
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
	c.rpl, err = tokens.NewTokenRpl(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating RPL token binding: %w", err)
	}
	c.odaoMgr, err = oracle.NewOracleDaoManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating Oracle DAO manager binding: %w", err)
	}
	c.oSettings = c.odaoMgr.Settings
	c.dnta, err = c.rp.GetContract(rocketpool.ContractName_RocketDAONodeTrustedActions)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting RPL token contract: %w", err)
	}
	return types.ResponseStatus_Success, nil
}

func (c *oracleDaoJoinContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.odaoMember.InvitedTime,
		c.oSettings.Proposal.ActionTime,
		c.odaoMember.Exists,
		c.oSettings.Member.RplBond,
	)
	c.rpl.GetAllowance(mc, &c.rplAllowance, c.nodeAddress, c.dnta.Address)
	c.rpl.BalanceOf(mc, &c.rplBalance, c.nodeAddress)
}

func (c *oracleDaoJoinContext) PrepareData(data *api.OracleDaoJoinData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	// Get the timestamp of the latest block
	ctx := c.handler.ctx
	latestHeader, err := c.rp.Client.HeaderByNumber(ctx, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting latest block header: %w", err)
	}
	currentTime := time.Unix(int64(latestHeader.Time), 0)
	actionWindow := c.oSettings.Proposal.ActionTime.Formatted()
	rplBond := c.oSettings.Member.RplBond.Get()

	// Check proposal details
	data.ProposalExpired = !isProposalActionable(actionWindow, c.odaoMember.InvitedTime.Formatted(), currentTime)
	data.AlreadyMember = c.odaoMember.Exists.Get()
	data.InsufficientRplBalance = (c.rplBalance.Cmp(rplBond) < 0)
	data.CanJoin = !(data.ProposalExpired || data.AlreadyMember || data.InsufficientRplBalance)

	// Get the tx
	if data.CanJoin {
		if c.rplAllowance.Cmp(rplBond) < 0 {
			// Do the approve TX if needed
			diff := big.NewInt(0).Sub(rplBond, c.rplAllowance)
			approveTxInfo, err := c.rpl.Approve(c.dnta.Address, diff, opts)
			if err != nil {
				return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for RPL approval: %w", err)
			}
			data.ApproveTxInfo = approveTxInfo
		} else {
			// Just do the join
			joinTxInfo, err := c.odaoMgr.Join(opts)
			if err != nil {
				return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for Join: %w", err)
			}
			data.JoinTxInfo = joinTxInfo
		}
	}
	return types.ResponseStatus_Success, nil
}
