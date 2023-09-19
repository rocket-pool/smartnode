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
	// Members
	c.oSettings.GetQuorum(mc)
	c.oSettings.GetRplBond(mc)
	c.oSettings.GetUnbondedMinipoolMax(mc)
	c.oSettings.GetChallengeCooldown(mc)
	c.oSettings.GetChallengeWindow(mc)
	c.oSettings.GetChallengeCost(mc)

	// Minipools
	c.oSettings.GetScrubPeriod(mc)
	c.oSettings.GetPromotionScrubPeriod(mc)
	c.oSettings.GetScrubPenaltyEnabled(mc)
	c.oSettings.GetBondReductionWindowStart(mc)
	c.oSettings.GetBondReductionWindowLength(mc)

	// Proposals
	c.oSettings.GetProposalCooldownTime(mc)
	c.oSettings.GetVoteTime(mc)
	c.oSettings.GetVoteDelayTime(mc)
	c.oSettings.GetProposalExecuteTime(mc)
	c.oSettings.GetProposalActionTime(mc)

}

func (c *oracleDaoSettingsContext) PrepareData(data *api.OracleDaoSettingsData, opts *bind.TransactOpts) error {
	data.Members.Quorum = c.oSettings.Members.Quorum.Formatted()
	data.Members.RplBond = c.oSettings.Members.RplBond
	data.Members.MinipoolUnbondedMax = c.oSettings.Members.UnbondedMinipoolMax.Formatted()
	data.Members.ChallengeCooldown = c.oSettings.Members.ChallengeCooldown.Formatted()
	data.Members.ChallengeWindow = c.oSettings.Members.ChallengeWindow.Formatted()
	data.Members.ChallengeCost = c.oSettings.Members.ChallengeCost

	data.Minipools.ScrubPeriod = c.oSettings.Minipools.ScrubPeriod.Formatted()
	data.Minipools.PromotionScrubPeriod = c.oSettings.Minipools.PromotionScrubPeriod.Formatted()
	data.Minipools.ScrubPenaltyEnabled = c.oSettings.Minipools.IsScrubPenaltyEnabled
	data.Minipools.BondReductionWindowStart = c.oSettings.Minipools.BondReductionWindowStart.Formatted()
	data.Minipools.BondReductionWindowLength = c.oSettings.Minipools.BondReductionWindowLength.Formatted()

	data.Proposals.Cooldown = c.oSettings.Proposals.CooldownTime.Formatted()
	data.Proposals.VoteTime = c.oSettings.Proposals.VoteTime.Formatted()
	data.Proposals.VoteDelayTime = c.oSettings.Proposals.VoteDelayTime.Formatted()
	data.Proposals.ExecuteTime = c.oSettings.Proposals.ExecuteTime.Formatted()
	data.Proposals.ActionTime = c.oSettings.Proposals.ActionTime.Formatted()
	return nil
}
