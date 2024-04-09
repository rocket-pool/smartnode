package odao

import (
	"fmt"
	"net/url"

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

type oracleDaoSettingsContextFactory struct {
	handler *OracleDaoHandler
}

func (f *oracleDaoSettingsContextFactory) Create(args url.Values) (*oracleDaoSettingsContext, error) {
	c := &oracleDaoSettingsContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *oracleDaoSettingsContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*oracleDaoSettingsContext, api.OracleDaoSettingsData](
		router, "settings", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type oracleDaoSettingsContext struct {
	handler     *OracleDaoHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	oSettings *oracle.OracleDaoSettings
}

func (c *oracleDaoSettingsContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	err := sp.RequireEthClientSynced(c.handler.ctx)
	if err != nil {
		return types.ResponseStatus_ClientsNotSynced, err
	}

	// Bindings
	odaoMgr, err := oracle.NewOracleDaoManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating oracle DAO manager binding: %w", err)
	}
	c.oSettings = odaoMgr.Settings
	return types.ResponseStatus_Success, nil
}

func (c *oracleDaoSettingsContext) GetState(mc *batch.MultiCaller) {
	eth.QueryAllFields(c.oSettings, mc)
}

func (c *oracleDaoSettingsContext) PrepareData(data *api.OracleDaoSettingsData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	data.Member.Quorum = c.oSettings.Member.Quorum.Formatted()
	data.Member.RplBond = c.oSettings.Member.RplBond.Get()
	data.Member.ChallengeCooldown = c.oSettings.Member.ChallengeCooldown.Formatted()
	data.Member.ChallengeWindow = c.oSettings.Member.ChallengeWindow.Formatted()
	data.Member.ChallengeCost = c.oSettings.Member.ChallengeCost.Get()

	data.Minipool.ScrubPeriod = c.oSettings.Minipool.ScrubPeriod.Formatted()
	data.Minipool.ScrubQuorum = c.oSettings.Minipool.ScrubQuorum.Formatted()
	data.Minipool.PromotionScrubPeriod = c.oSettings.Minipool.PromotionScrubPeriod.Formatted()
	data.Minipool.IsScrubPenaltyEnabled = c.oSettings.Minipool.IsScrubPenaltyEnabled.Get()
	data.Minipool.BondReductionWindowStart = c.oSettings.Minipool.BondReductionWindowStart.Formatted()
	data.Minipool.BondReductionWindowLength = c.oSettings.Minipool.BondReductionWindowLength.Formatted()
	data.Minipool.BondReductionCancellationQuorum = c.oSettings.Minipool.BondReductionCancellationQuorum.Formatted()

	data.Proposal.Cooldown = c.oSettings.Proposal.CooldownTime.Formatted()
	data.Proposal.VoteTime = c.oSettings.Proposal.VoteTime.Formatted()
	data.Proposal.VoteDelayTime = c.oSettings.Proposal.VoteDelayTime.Formatted()
	data.Proposal.ExecuteTime = c.oSettings.Proposal.ExecuteTime.Formatted()
	data.Proposal.ActionTime = c.oSettings.Proposal.ActionTime.Formatted()
	return types.ResponseStatus_Success, nil
}
