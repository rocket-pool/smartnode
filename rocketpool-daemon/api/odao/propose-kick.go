package odao

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/dao/oracle"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"

	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
	"github.com/rocket-pool/smartnode/shared/utils/math"
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
		router, "propose-kick", f, f.handler.serviceProvider,
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

func (c *oracleDaoProposeKickContext) Initialize() error {
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
	c.candidate, err = oracle.NewOracleDaoMember(c.rp, c.address)
	if err != nil {
		return fmt.Errorf("error creating candidate oracle DAO member binding: %w", err)
	}
	c.odaoMgr, err = oracle.NewOracleDaoManager(c.rp)
	if err != nil {
		return fmt.Errorf("error creating Oracle DAO manager binding: %w", err)
	}
	c.oSettings = c.odaoMgr.Settings
	return nil
}

func (c *oracleDaoProposeKickContext) GetState(mc *batch.MultiCaller) {
	core.AddQueryablesToMulticall(mc,
		c.odaoMember.LastProposalTime,
		c.candidate.Exists,
		c.candidate.RplBondAmount,
		c.candidate.ID,
		c.candidate.Url,
	)
	c.oSettings.Proposal.CooldownTime.AddToQuery(mc)
}

func (c *oracleDaoProposeKickContext) PrepareData(data *api.OracleDaoProposeKickData, opts *bind.TransactOpts) error {
	// Get the timestamp of the latest block
	latestHeader, err := c.rp.Client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("error getting latest block header: %w", err)
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
			return fmt.Errorf("error getting TX info for ProposeKickMember: %w", err)
		}
		data.TxInfo = txInfo
	}
	return nil
}
