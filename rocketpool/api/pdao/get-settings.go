package pdao

import (
	"time"

	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getSettings(c *cli.Context) (*api.GetPDAOSettingsResponse, error) {

	// Get services
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.GetPDAOSettingsResponse{}

	// Data
	var wg errgroup.Group

	// === Auction ===

	wg.Go(func() error {
		var err error
		response.Auction.IsCreateLotEnabled, err = protocol.GetCreateLotEnabled(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Auction.IsBidOnLotEnabled, err = protocol.GetBidOnLotEnabled(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Auction.LotMinimumEthValue, err = protocol.GetLotMinimumEthValue(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Auction.LotMaximumEthValue, err = protocol.GetLotMaximumEthValue(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Auction.LotDuration, err = protocol.GetLotDuration(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Auction.LotStartingPriceRatio, err = protocol.GetLotStartingPriceRatio(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Auction.LotReservePriceRatio, err = protocol.GetLotReservePriceRatio(rp, nil)
		return err
	})

	// === Deposit ===

	wg.Go(func() error {
		var err error
		response.Deposit.IsDepositingEnabled, err = protocol.GetDepositEnabled(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Deposit.AreDepositAssignmentsEnabled, err = protocol.GetAssignDepositsEnabled(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Deposit.MinimumDeposit, err = protocol.GetMinimumDeposit(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Deposit.MaximumDepositPoolSize, err = protocol.GetMaximumDepositPoolSize(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Deposit.MaximumAssignmentsPerDeposit, err = protocol.GetMaximumDepositAssignments(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Deposit.MaximumSocialisedAssignmentsPerDeposit, err = protocol.GetMaximumSocializedDepositAssignments(rp, nil)
		return err
	})

	wg.Go(func() error {
		depositFee, err := protocol.GetDepositFee(rp, nil)
		if err == nil {
			response.Deposit.DepositFee = eth.WeiToEth(depositFee)
		}
		return err
	})

	// === Inflation ===

	wg.Go(func() error {
		var err error
		response.Inflation.IntervalRate, err = protocol.GetInflationIntervalRate(rp, nil)
		return err
	})

	wg.Go(func() error {
		startTime, err := protocol.GetInflationStartTime(rp, nil)
		if err == nil {
			response.Inflation.StartTime = time.Unix(int64(startTime), 0)
		}
		return err
	})

	// === Minipool ===

	wg.Go(func() error {
		var err error
		response.Minipool.IsSubmitWithdrawableEnabled, err = protocol.GetMinipoolSubmitWithdrawableEnabled(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Minipool.LaunchTimeout, err = protocol.GetMinipoolLaunchTimeout(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Minipool.IsBondReductionEnabled, err = protocol.GetMinipoolSubmitWithdrawableEnabled(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Minipool.MaximumCount, err = protocol.GetMaximumMinipoolCount(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Minipool.UserDistributeWindowStart, err = protocol.GetMinipoolUserDistributeWindowStart(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Minipool.UserDistributeWindowLength, err = protocol.GetMinipoolUserDistributeWindowLength(rp, nil)
		return err
	})

	// === Network ===

	wg.Go(func() error {
		var err error
		response.Network.OracleDaoConsensusThreshold, err = protocol.GetNodeConsensusThreshold(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Network.NodePenaltyThreshold, err = protocol.GetNetworkPenaltyThreshold(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Network.PerPenaltyRate, err = protocol.GetNetworkPenaltyPerRate(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Network.IsSubmitBalancesEnabled, err = protocol.GetSubmitBalancesEnabled(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Network.SubmitBalancesFrequency, err = protocol.GetSubmitBalancesFrequency(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Network.IsSubmitPricesEnabled, err = protocol.GetSubmitPricesEnabled(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Network.SubmitPricesFrequency, err = protocol.GetSubmitPricesFrequency(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Network.MinimumNodeFee, err = protocol.GetMinimumNodeFee(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Network.TargetNodeFee, err = protocol.GetTargetNodeFee(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Network.MaximumNodeFee, err = protocol.GetMaximumNodeFee(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Network.NodeFeeDemandRange, err = protocol.GetNodeFeeDemandRange(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Network.TargetRethCollateralRate, err = protocol.GetTargetRethCollateralRate(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Network.IsSubmitRewardsEnabled, err = protocol.GetSubmitRewardsEnabled(rp, nil)
		return err
	})

	// === Node ===

	wg.Go(func() error {
		var err error
		response.Node.IsRegistrationEnabled, err = protocol.GetNodeRegistrationEnabled(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Node.IsSmoothingPoolRegistrationEnabled, err = protocol.GetSmoothingPoolRegistrationEnabled(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Node.IsDepositingEnabled, err = protocol.GetNodeDepositEnabled(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Node.AreVacantMinipoolsEnabled, err = protocol.GetVacantMinipoolsEnabled(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Node.MinimumPerMinipoolStake, err = protocol.GetMinimumPerMinipoolStake(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Node.MaximumPerMinipoolStake, err = protocol.GetMaximumPerMinipoolStake(rp, nil)
		return err
	})

	// === Proposals ===

	wg.Go(func() error {
		var err error
		response.Proposals.VotePhase1Time, err = protocol.GetVotePhase1Time(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Proposals.VotePhase2Time, err = protocol.GetVotePhase2Time(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Proposals.VoteDelayTime, err = protocol.GetVoteDelayTime(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Proposals.ExecuteTime, err = protocol.GetExecuteTime(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Proposals.ProposalBond, err = protocol.GetProposalBond(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Proposals.ChallengeBond, err = protocol.GetChallengeBond(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Proposals.ChallengePeriod, err = protocol.GetChallengePeriod(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Proposals.Quorum, err = protocol.GetProposalQuorum(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Proposals.VetoQuorum, err = protocol.GetProposalVetoQuorum(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Proposals.MaxBlockAge, err = protocol.GetProposalMaxBlockAge(rp, nil)
		return err
	})

	// === Rewards ===

	wg.Go(func() error {
		var err error
		response.Rewards.IntervalTime, err = protocol.GetRewardsClaimIntervalTime(rp, nil)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Return response
	return &response, nil
}
