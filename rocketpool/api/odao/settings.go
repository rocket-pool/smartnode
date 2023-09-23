package odao

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
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
	// Member
	c.oSettings.Member.Quorum.Get(mc)
	c.oSettings.Member.RplBond.Get(mc)
	c.oSettings.Member.UnbondedMinipoolMax.Get(mc)
	c.oSettings.Member.ChallengeCooldown.Get(mc)
	c.oSettings.Member.ChallengeWindow.Get(mc)
	c.oSettings.Member.ChallengeCost.Get(mc)

	// Minipool
	c.oSettings.Minipool.ScrubPeriod.Get(mc)
	c.oSettings.Minipool.PromotionScrubPeriod.Get(mc)
	c.oSettings.Minipool.IsScrubPenaltyEnabled.Get(mc)
	c.oSettings.Minipool.BondReductionWindowStart.Get(mc)
	c.oSettings.Minipool.BondReductionWindowLength.Get(mc)

	// Proposal
	c.oSettings.Proposal.CooldownTime.Get(mc)
	c.oSettings.Proposal.VoteTime.Get(mc)
	c.oSettings.Proposal.VoteDelayTime.Get(mc)
	c.oSettings.Proposal.ExecuteTime.Get(mc)
	c.oSettings.Proposal.ActionTime.Get(mc)

}

func (c *oracleDaoSettingsContext) PrepareData(data *api.OracleDaoSettingsData, opts *bind.TransactOpts) error {
	data.Members.Quorum = c.oSettings.Member.Quorum.Value.Formatted()
	data.Members.RplBond = c.oSettings.Member.RplBond.Value
	data.Members.MinipoolUnbondedMax = c.oSettings.Member.UnbondedMinipoolMax.Value.Formatted()
	data.Members.ChallengeCooldown = c.oSettings.Member.ChallengeCooldown.Value.Formatted()
	data.Members.ChallengeWindow = c.oSettings.Member.ChallengeWindow.Value.Formatted()
	data.Members.ChallengeCost = c.oSettings.Member.ChallengeCost.Value

	data.Minipools.ScrubPeriod = c.oSettings.Minipool.ScrubPeriod.Value.Formatted()
	data.Minipools.PromotionScrubPeriod = c.oSettings.Minipool.PromotionScrubPeriod.Value.Formatted()
	data.Minipools.ScrubPenaltyEnabled = c.oSettings.Minipool.IsScrubPenaltyEnabled.Value
	data.Minipools.BondReductionWindowStart = c.oSettings.Minipool.BondReductionWindowStart.Value.Formatted()
	data.Minipools.BondReductionWindowLength = c.oSettings.Minipool.BondReductionWindowLength.Value.Formatted()

	data.Proposals.Cooldown = c.oSettings.Proposal.CooldownTime.Value.Formatted()
	data.Proposals.VoteTime = c.oSettings.Proposal.VoteTime.Value.Formatted()
	data.Proposals.VoteDelayTime = c.oSettings.Proposal.VoteDelayTime.Value.Formatted()
	data.Proposals.ExecuteTime = c.oSettings.Proposal.ExecuteTime.Value.Formatted()
	data.Proposals.ActionTime = c.oSettings.Proposal.ActionTime.Value.Formatted()
	return nil
}
