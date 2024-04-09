package odao

import (
	"errors"
	"fmt"
	"math/big"
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
	"github.com/rocket-pool/node-manager-core/utils/math"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type oracleDaoProposeKickContextFactory struct {
	handler *OracleDaoHandler
}

func (f *oracleDaoProposeKickContextFactory) Create(args url.Values) (*oracleDaoProposeKickContext, error) {
	c := &oracleDaoProposeKickContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("address", args, input.ValidateAddress, &c.address),
		server.ValidateArg("fineAmount", args, input.ValidateBigInt, &c.fineAmount),
	}
	return c, errors.Join(inputErrs...)
}

func (f *oracleDaoProposeKickContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*oracleDaoProposeKickContext, api.OracleDaoProposeKickData](
		router, "propose-kick", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type oracleDaoProposeKickContext struct {
	handler     *OracleDaoHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	address    common.Address
	fineAmount *big.Int
	odaoMember *oracle.OracleDaoMember
	candidate  *oracle.OracleDaoMember
	oSettings  *oracle.OracleDaoSettings
	odaoMgr    *oracle.OracleDaoManager
}

func (c *oracleDaoProposeKickContext) Initialize() (types.ResponseStatus, error) {
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
	c.candidate, err = oracle.NewOracleDaoMember(c.rp, c.address)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating candidate oracle DAO member binding: %w", err)
	}
	c.odaoMgr, err = oracle.NewOracleDaoManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating Oracle DAO manager binding: %w", err)
	}
	c.oSettings = c.odaoMgr.Settings
	return types.ResponseStatus_Success, nil
}

func (c *oracleDaoProposeKickContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.odaoMember.LastProposalTime,
		c.candidate.Exists,
		c.candidate.RplBondAmount,
		c.candidate.ID,
		c.candidate.Url,
	)
	c.oSettings.Proposal.CooldownTime.AddToQuery(mc)
}

func (c *oracleDaoProposeKickContext) PrepareData(data *api.OracleDaoProposeKickData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	// Get the timestamp of the latest block
	ctx := c.handler.ctx
	latestHeader, err := c.rp.Client.HeaderByNumber(ctx, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting latest block header: %w", err)
	}
	currentTime := time.Unix(int64(latestHeader.Time), 0)
	cooldownTime := c.oSettings.Proposal.CooldownTime.Formatted()

	// Check proposal details
	data.MemberDoesNotExist = !c.candidate.Exists.Get()
	data.ProposalCooldownActive = isProposalCooldownActive(cooldownTime, c.odaoMember.LastProposalTime.Formatted(), currentTime)
	data.InsufficientRplBond = (c.fineAmount.Cmp(c.candidate.RplBondAmount.Get()) > 0)
	data.CanPropose = !(data.MemberDoesNotExist || data.ProposalCooldownActive || data.InsufficientRplBond)

	// Get the tx
	if data.CanPropose && opts != nil {
		message := fmt.Sprintf("kick %s (%s) with %.6f RPL fine", c.candidate.ID.Get(), c.candidate.Url.Get(), math.RoundDown(eth.WeiToEth(c.fineAmount), 6))
		txInfo, err := c.odaoMgr.ProposeKickMember(message, c.address, c.fineAmount, opts)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for ProposeKickMember: %w", err)
		}
		data.TxInfo = txInfo
	}
	return types.ResponseStatus_Success, nil
}
