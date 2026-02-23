package pdao

import (
	"time"

	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	protocol131 "github.com/rocket-pool/smartnode/bindings/legacy/v1.3.1/protocol"
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

	// New calls introduced in Saturn
	if response.SaturnDeployed {

		// === Deposit ===

		wg.Go(func() error {
			var err error
			response.Deposit.ExpressQueueRate, err = protocol.GetExpressQueueRate(rp, nil)
			return err
		})

		wg.Go(func() error {
			var err error
			response.Deposit.ExpressQueueTicketsBaseProvision, err = protocol.GetExpressQueueTicketsBaseProvision(rp, nil)
			return err
		})

		// === Minipool ===

		// MaximumPenaltyCount is not live on devnet yet
		// wg.Go(func() error {
		// 	var err error
		// 	response.Minipool.MaximumPenaltyCount, err = protocol.GetMaximumPenaltyCount(rp, nil)
		// 	return err
		// })

		// === Network ===

		wg.Go(func() error {
			var err error
			response.Network.NodeCommissionShare, err = protocol.GetNodeShare(rp, nil)
			return err
		})

		wg.Go(func() error {
			var err error
			response.Network.NodeCommissionShareSecurityCouncilAdder, err = protocol.GetNodeShareSecurityCouncilAdder(rp, nil)
			return err
		})

		wg.Go(func() error {
			var err error
			response.Network.VoterShare, err = protocol.GetVoterShare(rp, nil)
			return err
		})

		wg.Go(func() error {
			var err error
			response.Network.ProtocolDAOShare, err = protocol.GetProtocolDAOShare(rp, nil)
			return err
		})

		wg.Go(func() error {
			var err error
			response.Network.MaxNodeShareSecurityCouncilAdder, err = protocol.GetMaxNodeShareSecurityCouncilAdder(rp, nil)
			return err
		})

		wg.Go(func() error {
			var err error
			response.Network.MaxRethBalanceDelta, err = protocol.GetMaxRethDelta(rp, nil)
			return err
		})

		wg.Go(func() error {
			var err error
			response.Network.AllowListedControllers, err = protocol.GetAllowListedControllers(rp, nil)
			return err
		})

		// === Node ===

		wg.Go(func() error {
			var err error
			response.Node.ReducedBond, err = protocol.GetReducedBond(rp, nil)
			return err
		})

		wg.Go(func() error {
			var err error
			nodeUnstakingPeriod, err := protocol.GetNodeUnstakingPeriod(rp, nil)
			if err == nil {
				response.Node.NodeUnstakingPeriod = time.Duration(nodeUnstakingPeriod.Int64()) * time.Second
			}
			return err
		})

		wg.Go(func() error {
			var err error
			response.Node.MinimumLegacyRplStake, err = protocol.GetMinimumLegacyRPLStakeRaw(rp, nil)
			return err
		})

		// === Megapool ===

		wg.Go(func() error {
			var err error
			timeBeforeDissolve, err := protocol.GetMegapoolTimeBeforeDissolve(rp, nil)
			if err == nil {
				response.Megapool.TimeBeforeDissolve = time.Duration(timeBeforeDissolve) * time.Second
			}
			return err
		})

		wg.Go(func() error {
			var err error
			response.Megapool.MaximumEthPenalty, err = protocol.GetMaximumEthPenalty(rp, nil)
			return err
		})

		wg.Go(func() error {
			var err error
			response.Megapool.NotifyThreshold, err = protocol.GetNotifyThreshold(rp, nil)
			return err
		})

		wg.Go(func() error {
			var err error
			response.Megapool.LateNotifyFine, err = protocol.GetLateNotifyFine(rp, nil)
			return err
		})

		wg.Go(func() error {
			var err error
			response.Megapool.DissolvePenalty, err = protocol.GetMegapoolDissolvePenalty(rp, nil)
			return err
		})

		wg.Go(func() error {
			var err error
			response.Megapool.UserDistributeDelay, err = protocol.GetUserDistributeDelay(rp, nil)
			return err
		})

		wg.Go(func() error {
			var err error
			response.Megapool.UserDistributeDelayWithShortfall, err = protocol.GetUserDistributeDelayWithShortfall(rp, nil)
			return err
		})

		wg.Go(func() error {
			var err error
			response.Megapool.PenaltyThreshold, err = protocol.GetPenaltyThreshold(rp, nil)
			return err
		})

	}

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
		response.Auction.LotStartingPriceRatio, err = protocol.GetLotStartingPriceRatioRaw(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Auction.LotReservePriceRatio, err = protocol.GetLotReservePriceRatioRaw(rp, nil)
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
			response.Deposit.DepositFee = depositFee
		}
		return err
	})

	// === Inflation ===

	wg.Go(func() error {
		var err error
		response.Inflation.IntervalRate, err = protocol.GetInflationIntervalRateRaw(rp, nil)
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
		response.Network.OracleDaoConsensusThreshold, err = protocol.GetNodeConsensusThresholdRaw(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Network.NodePenaltyThreshold, err = protocol.GetNetworkPenaltyThresholdRaw(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Network.PerPenaltyRate, err = protocol.GetNetworkPenaltyPerRateRaw(rp, nil)
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
		response.Network.MinimumNodeFee, err = protocol.GetMinimumNodeFeeRaw(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Network.TargetNodeFee, err = protocol.GetTargetNodeFeeRaw(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Network.MaximumNodeFee, err = protocol.GetMaximumNodeFeeRaw(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Network.NodeFeeDemandRange, err = protocol.GetNodeFeeDemandRange(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Network.TargetRethCollateralRate, err = protocol.GetTargetRethCollateralRateRaw(rp, nil)
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

	// In Saturn, these two bindings are deprecated in favor of 'GetMinimumLegacyRPLStake'
	if !response.SaturnDeployed {
		wg.Go(func() error {
			var err error
			response.Node.MinimumPerMinipoolStake, err = protocol131.GetMinimumPerMinipoolStakeRaw(rp, nil)
			return err
		})

		wg.Go(func() error {
			var err error
			response.Node.MaximumPerMinipoolStake, err = protocol131.GetMaximumPerMinipoolStakeRaw(rp, nil)
			return err
		})
	}

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
		response.Proposals.Quorum, err = protocol.GetProposalQuorumRaw(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Proposals.VetoQuorum, err = protocol.GetProposalVetoQuorumRaw(rp, nil)
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

	// === Security ===

	wg.Go(func() error {
		var err error
		response.Security.MembersQuorum, err = protocol.GetSecurityMembersQuorum(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Security.MembersLeaveTime, err = protocol.GetSecurityMembersLeaveTime(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Security.ProposalVoteTime, err = protocol.GetSecurityProposalVoteTime(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Security.ProposalExecuteTime, err = protocol.GetSecurityProposalExecuteTime(rp, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		response.Security.ProposalActionTime, err = protocol.GetSecurityProposalActionTime(rp, nil)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Return response
	return &response, nil
}
