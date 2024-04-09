package pdao

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type protocolDaoSettingsContextFactory struct {
	handler *ProtocolDaoHandler
}

func (f *protocolDaoSettingsContextFactory) Create(args url.Values) (*protocolDaoSettingsContext, error) {
	c := &protocolDaoSettingsContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *protocolDaoSettingsContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*protocolDaoSettingsContext, api.ProtocolDaoSettingsData](
		router, "settings", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type protocolDaoSettingsContext struct {
	handler *ProtocolDaoHandler
	rp      *rocketpool.RocketPool

	pMgr      *protocol.ProtocolDaoManager
	pSettings *protocol.ProtocolDaoSettings
}

func (c *protocolDaoSettingsContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()

	// Requirements
	status, err := sp.RequireRocketPoolContracts(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	c.pMgr, err = protocol.NewProtocolDaoManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating protocol DAO manager binding: %w", err)
	}
	c.pSettings = c.pMgr.Settings
	return types.ResponseStatus_Success, nil
}

func (c *protocolDaoSettingsContext) GetState(mc *batch.MultiCaller) {
	eth.QueryAllFields(c.pSettings, mc)
	eth.AddQueryablesToMulticall(mc,
		c.pMgr.IntervalTime,
	)
}

func (c *protocolDaoSettingsContext) PrepareData(data *api.ProtocolDaoSettingsData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	data.Auction.IsCreateLotEnabled = c.pSettings.Auction.IsCreateLotEnabled.Get()
	data.Auction.IsBidOnLotEnabled = c.pSettings.Auction.IsBidOnLotEnabled.Get()
	data.Auction.LotMinimumEthValue = c.pSettings.Auction.LotMinimumEthValue.Get()
	data.Auction.LotMaximumEthValue = c.pSettings.Auction.LotMaximumEthValue.Get()
	data.Auction.LotDuration = c.pSettings.Auction.LotDuration.Formatted()
	data.Auction.LotStartingPriceRatio = c.pSettings.Auction.LotStartingPriceRatio.Raw()
	data.Auction.LotReservePriceRatio = c.pSettings.Auction.LotReservePriceRatio.Raw()

	data.Deposit.IsDepositingEnabled = c.pSettings.Deposit.IsDepositingEnabled.Get()
	data.Deposit.AreDepositAssignmentsEnabled = c.pSettings.Deposit.AreDepositAssignmentsEnabled.Get()
	data.Deposit.MinimumDeposit = c.pSettings.Deposit.MinimumDeposit.Get()
	data.Deposit.MaximumDepositPoolSize = c.pSettings.Deposit.MaximumDepositPoolSize.Get()
	data.Deposit.MaximumAssignmentsPerDeposit = c.pSettings.Deposit.MaximumAssignmentsPerDeposit.Formatted()
	data.Deposit.MaximumSocialisedAssignmentsPerDeposit = c.pSettings.Deposit.MaximumSocialisedAssignmentsPerDeposit.Formatted()
	data.Deposit.DepositFee = c.pSettings.Deposit.DepositFee.Raw()

	data.Inflation.IntervalRate = c.pSettings.Inflation.IntervalRate.Raw()
	data.Inflation.StartTime = c.pSettings.Inflation.StartTime.Formatted()

	data.Minipool.IsSubmitWithdrawableEnabled = c.pSettings.Minipool.IsSubmitWithdrawableEnabled.Get()
	data.Minipool.LaunchTimeout = c.pSettings.Minipool.LaunchTimeout.Formatted()
	data.Minipool.IsBondReductionEnabled = c.pSettings.Minipool.IsBondReductionEnabled.Get()
	data.Minipool.MaximumCount = c.pSettings.Minipool.MaximumCount.Formatted()
	data.Minipool.UserDistributeWindowStart = c.pSettings.Minipool.UserDistributeWindowStart.Formatted()
	data.Minipool.UserDistributeWindowLength = c.pSettings.Minipool.UserDistributeWindowLength.Formatted()

	data.Network.OracleDaoConsensusThreshold = c.pSettings.Network.OracleDaoConsensusThreshold.Raw()
	data.Network.NodePenaltyThreshold = c.pSettings.Network.NodePenaltyThreshold.Raw()
	data.Network.PerPenaltyRate = c.pSettings.Network.PerPenaltyRate.Raw()
	data.Network.IsSubmitBalancesEnabled = c.pSettings.Network.IsSubmitBalancesEnabled.Get()
	data.Network.SubmitBalancesFrequency = c.pSettings.Network.SubmitBalancesFrequency.Formatted()
	data.Network.IsSubmitPricesEnabled = c.pSettings.Network.IsSubmitPricesEnabled.Get()
	data.Network.SubmitPricesFrequency = c.pSettings.Network.SubmitPricesFrequency.Formatted()
	data.Network.MinimumNodeFee = c.pSettings.Network.MinimumNodeFee.Raw()
	data.Network.TargetNodeFee = c.pSettings.Network.TargetNodeFee.Raw()
	data.Network.MaximumNodeFee = c.pSettings.Network.MaximumNodeFee.Raw()
	data.Network.NodeFeeDemandRange = c.pSettings.Network.NodeFeeDemandRange.Get()
	data.Network.TargetRethCollateralRate = c.pSettings.Network.TargetRethCollateralRate.Raw()
	data.Network.IsSubmitRewardsEnabled = c.pSettings.Network.IsSubmitRewardsEnabled.Get()

	data.Node.IsRegistrationEnabled = c.pSettings.Node.IsRegistrationEnabled.Get()
	data.Node.IsSmoothingPoolRegistrationEnabled = c.pSettings.Node.IsSmoothingPoolRegistrationEnabled.Get()
	data.Node.IsDepositingEnabled = c.pSettings.Node.IsDepositingEnabled.Get()
	data.Node.AreVacantMinipoolsEnabled = c.pSettings.Node.AreVacantMinipoolsEnabled.Get()
	data.Node.MinimumPerMinipoolStake = c.pSettings.Node.MinimumPerMinipoolStake.Raw()
	data.Node.MaximumPerMinipoolStake = c.pSettings.Node.MaximumPerMinipoolStake.Raw()

	data.Proposals.VotePhase1Time = c.pSettings.Proposals.VotePhase1Time.Formatted()
	data.Proposals.VotePhase2Time = c.pSettings.Proposals.VotePhase2Time.Formatted()
	data.Proposals.VoteDelayTime = c.pSettings.Proposals.VoteDelayTime.Formatted()
	data.Proposals.ExecuteTime = c.pSettings.Proposals.ExecuteTime.Formatted()
	data.Proposals.ProposalBond = c.pSettings.Proposals.ProposalBond.Get()
	data.Proposals.ChallengeBond = c.pSettings.Proposals.ChallengeBond.Get()
	data.Proposals.Quorum = c.pSettings.Proposals.ProposalQuorum.Raw()
	data.Proposals.VetoQuorum = c.pSettings.Proposals.ProposalVetoQuorum.Raw()
	data.Proposals.MaxBlockAge = c.pSettings.Proposals.ProposalMaxBlockAge.Formatted()

	data.Rewards.IntervalTime = c.pMgr.IntervalTime.Formatted()

	data.Security.MembersQuorum = c.pSettings.Security.MembersQuorum.Raw()
	data.Security.MembersLeaveTime = c.pSettings.Security.MembersLeaveTime.Formatted()
	data.Security.ProposalVoteTime = c.pSettings.Security.ProposalVoteTime.Formatted()
	data.Security.ProposalExecuteTime = c.pSettings.Security.ProposalExecuteTime.Formatted()
	data.Security.ProposalActionTime = c.pSettings.Security.ProposalActionTime.Formatted()

	return types.ResponseStatus_Success, nil
}
