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
	"github.com/rocket-pool/smartnode/v2/shared/utils"
)

// ===============
// === Factory ===
// ===============

type oracleDaoProposeInviteContextFactory struct {
	handler *OracleDaoHandler
}

func (f *oracleDaoProposeInviteContextFactory) Create(args url.Values) (*oracleDaoProposeInviteContext, error) {
	c := &oracleDaoProposeInviteContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("address", args, input.ValidateAddress, &c.address),
		server.ValidateArg("id", args, utils.ValidateDaoMemberID, &c.id),
		server.GetStringFromVars("url", args, &c.url),
	}
	return c, errors.Join(inputErrs...)
}

func (f *oracleDaoProposeInviteContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*oracleDaoProposeInviteContext, api.OracleDaoProposeInviteData](
		router, "propose-invite", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type oracleDaoProposeInviteContext struct {
	handler     *OracleDaoHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	address    common.Address
	id         string
	url        string
	odaoMember *oracle.OracleDaoMember
	candidate  *oracle.OracleDaoMember
	oSettings  *oracle.OracleDaoSettings
	odaoMgr    *oracle.OracleDaoManager
}

func (c *oracleDaoProposeInviteContext) Initialize() (types.ResponseStatus, error) {
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

func (c *oracleDaoProposeInviteContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.odaoMember.LastProposalTime,
		c.oSettings.Proposal.CooldownTime,
		c.candidate.Exists,
	)
}

func (c *oracleDaoProposeInviteContext) PrepareData(data *api.OracleDaoProposeInviteData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
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
	data.MemberAlreadyExists = c.candidate.Exists.Get()
	data.CanPropose = !(data.ProposalCooldownActive || data.MemberAlreadyExists)

	// Get the tx
	if data.CanPropose && opts != nil {
		message := fmt.Sprintf("invite %s (%s)", c.id, c.url)
		txInfo, err := c.odaoMgr.ProposeInviteMember(message, c.address, c.id, c.url, opts)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for ProposeInviteMember: %w", err)
		}
		data.TxInfo = txInfo
	}
	return types.ResponseStatus_Success, nil
}
