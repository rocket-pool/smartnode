package odao

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/dao/oracle"
	"github.com/rocket-pool/rocketpool-go/rocketpool"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type oracleDaoSettingsContextFactory struct {
	handler *OracleDaoHandler
}

func (f *oracleDaoSettingsContextFactory) Create(vars map[string]string) (*oracleDaoSettingsContext, error) {
	c := &oracleDaoSettingsContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *oracleDaoSettingsContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*oracleDaoSettingsContext, api.OracleDaoSettingsData](
		router, "settings", f, f.handler.serviceProvider,
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

func (c *oracleDaoSettingsContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	err := sp.RequireEthClientSynced()
	if err != nil {
		return err
	}

	// Bindings
	odaoMgr, err := oracle.NewOracleDaoManager(c.rp)
	if err != nil {
		return fmt.Errorf("error creating oracle DAO manager binding: %w", err)
	}
	c.oSettings = odaoMgr.Settings
	return nil
}

func (c *oracleDaoSettingsContext) GetState(mc *batch.MultiCaller) {
	core.QueryAllFields(c.oSettings, mc)
}

func (c *oracleDaoSettingsContext) PrepareData(data *api.OracleDaoSettingsData, opts *bind.TransactOpts) error {
	data.Members.Quorum = c.oSettings.Member.Quorum.Formatted()
	data.Members.RplBond = c.oSettings.Member.RplBond.Get()
	data.Members.MinipoolUnbondedMax = c.oSettings.Member.UnbondedMinipoolMax.Formatted()
	data.Members.ChallengeCooldown = c.oSettings.Member.ChallengeCooldown.Formatted()
	data.Members.ChallengeWindow = c.oSettings.Member.ChallengeWindow.Formatted()
	data.Members.ChallengeCost = c.oSettings.Member.ChallengeCost.Get()

	data.Minipools.ScrubPeriod = c.oSettings.Minipool.ScrubPeriod.Formatted()
	data.Minipools.PromotionScrubPeriod = c.oSettings.Minipool.PromotionScrubPeriod.Formatted()
	data.Minipools.ScrubPenaltyEnabled = c.oSettings.Minipool.IsScrubPenaltyEnabled.Get()
	data.Minipools.BondReductionWindowStart = c.oSettings.Minipool.BondReductionWindowStart.Formatted()
	data.Minipools.BondReductionWindowLength = c.oSettings.Minipool.BondReductionWindowLength.Formatted()

	data.Proposals.Cooldown = c.oSettings.Proposal.CooldownTime.Formatted()
	data.Proposals.VoteTime = c.oSettings.Proposal.VoteTime.Formatted()
	data.Proposals.VoteDelayTime = c.oSettings.Proposal.VoteDelayTime.Formatted()
	data.Proposals.ExecuteTime = c.oSettings.Proposal.ExecuteTime.Formatted()
	data.Proposals.ActionTime = c.oSettings.Proposal.ActionTime.Formatted()
	return nil
}
