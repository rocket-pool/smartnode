package pdao

import (
	"fmt"
	"math/big"
	"time"

	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

func proposeSettingAuctionIsCreateLotEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.CreateLotEnabledSettingPath, trueValue)
}

func proposeSettingAuctionIsBidOnLotEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.BidOnLotEnabledSettingPath, trueValue)
}

func proposeSettingAuctionLotMinimumEthValue(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.LotMinimumEthValueSettingPath, trueValue)
}

func proposeSettingAuctionLotMaximumEthValue(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.LotMaximumEthValueSettingPath, trueValue)
}

func proposeSettingAuctionLotDuration(c *cli.Context, value uint64) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.LotDurationSettingPath, trueValue)
}

func proposeSettingAuctionLotStartingPriceRatio(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.LotStartingPriceRatioSettingPath, trueValue)
}

func proposeSettingAuctionLotReservePriceRatio(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.LotReservePriceRatioSettingPath, trueValue)
}

func proposeSettingDepositIsDepositingEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.DepositEnabledSettingPath, trueValue)
}

func proposeSettingDepositAreDepositAssignmentsEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.AssignDepositsEnabledSettingPath, trueValue)
}

func proposeSettingDepositMinimumDeposit(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.MinimumDepositSettingPath, trueValue)
}

func proposeSettingDepositMaximumDepositPoolSize(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.MaximumDepositPoolSizeSettingPath, trueValue)
}

func proposeSettingDepositMaximumAssignmentsPerDeposit(c *cli.Context, value uint64) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.MaximumDepositAssignmentsSettingPath, trueValue)
}

func proposeSettingDepositMaximumSocialisedAssignmentsPerDeposit(c *cli.Context, value uint64) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.MaximumSocializedDepositAssignmentsSettingPath, trueValue)
}

func proposeSettingDepositDepositFee(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.DepositFeeSettingPath, trueValue)
}

func proposeSettingMinipoolIsSubmitWithdrawableEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.MinipoolSubmitWithdrawableEnabledSettingPath, trueValue)
}

func proposeSettingMinipoolLaunchTimeout(c *cli.Context, value time.Duration) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(c, protocol.MinipoolLaunchTimeoutSettingPath, trueValue)
}

func proposeSettingMinipoolIsBondReductionEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.BondReductionEnabledSettingPath, trueValue)
}

func proposeSettingMinipoolMaximumCount(c *cli.Context, value uint64) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.MaximumMinipoolCountSettingPath, trueValue)
}

func proposeSettingMinipoolUserDistributeWindowStart(c *cli.Context, value time.Duration) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(c, protocol.MinipoolUserDistributeWindowStartSettingPath, trueValue)
}

func proposeSettingMinipoolUserDistributeWindowLength(c *cli.Context, value time.Duration) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(c, protocol.MinipoolUserDistributeWindowLengthSettingPath, trueValue)
}

func proposeSettingNetworkOracleDaoConsensusThreshold(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.NodeConsensusThresholdSettingPath, trueValue)
}

func proposeSettingNetworkNodePenaltyThreshold(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.NetworkPenaltyThresholdSettingPath, trueValue)
}

func proposeSettingNetworkPerPenaltyRate(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.NetworkPenaltyPerRateSettingPath, trueValue)
}

func proposeSettingNetworkIsSubmitBalancesEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.SubmitBalancesEnabledSettingPath, trueValue)
}

func proposeSettingNetworkSubmitBalancesEpochs(c *cli.Context, value uint64) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.SubmitBalancesEpochsSettingPath, trueValue)
}

func proposeSettingNetworkIsSubmitPricesEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.SubmitPricesEnabledSettingPath, trueValue)
}

func proposeSettingNetworkSubmitPricesEpochs(c *cli.Context, value uint64) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.SubmitPricesEpochsSettingPath, trueValue)
}

func proposeSettingNetworkMinimumNodeFee(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.MinimumNodeFeeSettingPath, trueValue)
}

func proposeSettingNetworkTargetNodeFee(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.TargetNodeFeeSettingPath, trueValue)
}

func proposeSettingNetworkMaximumNodeFee(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.MaximumNodeFeeSettingPath, trueValue)
}

func proposeSettingNetworkNodeFeeDemandRange(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.NodeFeeDemandRangeSettingPath, trueValue)
}

func proposeSettingNetworkTargetRethCollateralRate(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.TargetRethCollateralRateSettingPath, trueValue)
}

func proposeSettingNetworkIsSubmitRewardsEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.SubmitRewardsEnabledSettingPath, trueValue)
}

func proposeSettingNodeIsRegistrationEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.NodeRegistrationEnabledSettingPath, trueValue)
}

func proposeSettingNodeIsSmoothingPoolRegistrationEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.SmoothingPoolRegistrationEnabledSettingPath, trueValue)
}

func proposeSettingNodeIsDepositingEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.NodeDepositEnabledSettingPath, trueValue)
}

func proposeSettingNodeAreVacantMinipoolsEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.VacantMinipoolsEnabledSettingPath, trueValue)
}

func proposeSettingNodeMinimumPerMinipoolStake(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.MinimumPerMinipoolStakeSettingPath, trueValue)
}

func proposeSettingNodeMaximumPerMinipoolStake(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.MaximumPerMinipoolStakeSettingPath, trueValue)
}

func proposeSettingProposalsVoteTime(c *cli.Context, value time.Duration) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(c, protocol.VoteTimeSettingPath, trueValue)
}

func proposeSettingProposalsVoteDelayTime(c *cli.Context, value time.Duration) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(c, protocol.VoteDelayTimeSettingPath, trueValue)
}

func proposeSettingProposalsExecuteTime(c *cli.Context, value time.Duration) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(c, protocol.ExecuteTimeSettingPath, trueValue)
}

func proposeSettingProposalsProposalBond(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.ProposalBondSettingPath, trueValue)
}

func proposeSettingProposalsChallengeBond(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.ChallengeBondSettingPath, trueValue)
}

func proposeSettingProposalsChallengePeriod(c *cli.Context, value time.Duration) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(c, protocol.ChallengePeriodSettingPath, trueValue)
}

func proposeSettingProposalsQuorum(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.ProposalQuorumSettingPath, trueValue)
}

func proposeSettingProposalsVetoQuorum(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.ProposalVetoQuorumSettingPath, trueValue)
}

func proposeSettingProposalsMaxBlockAge(c *cli.Context, value uint64) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.ProposalMaxBlockAgeSettingPath, trueValue)
}

func proposeSettingRewardsIntervalTime(c *cli.Context, value time.Duration) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(c, protocol.RewardsClaimIntervalTimeSettingPath, trueValue)
}

// Master general proposal function
func proposeSetting(c *cli.Context, setting string, value string) error {
	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check for Houston
	houston, err := rp.IsHoustonDeployed()
	if err != nil {
		return fmt.Errorf("error checking if Houston has been deployed: %w", err)
	}
	if !houston.IsHoustonDeployed {
		fmt.Println("This command cannot be used until Houston has been deployed.")
		return nil
	}

	// Check if proposal can be made
	canPropose, err := rp.PDAOCanProposeSetting(setting, value)
	if err != nil {
		return err
	}
	if !canPropose.CanPropose {
		fmt.Println("Cannot propose setting update:")
		if canPropose.InsufficientRpl {
			fmt.Printf("You do not have enough RPL staked but unlocked to make another proposal (unlocked: %.6f RPL, required: %.6f RPL).\n",
				eth.WeiToEth(big.NewInt(0).Sub(canPropose.StakedRpl, canPropose.LockedRpl)), eth.WeiToEth(canPropose.ProposalBond),
			)
		}
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canPropose.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm("Are you sure you want to submit this proposal?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Submit proposal
	response, err := rp.PDAOProposeSetting(setting, value, canPropose.BlockNumber, canPropose.Pollard)
	if err != nil {
		return err
	}

	fmt.Printf("Submitting proposal...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully submitted a %s setting update proposal with ID %d.\n", setting, response.ProposalId)
	return nil
}
