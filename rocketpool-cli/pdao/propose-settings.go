package pdao

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

func proposeSettingAuctionIsCreateLotEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.AuctionSettingsContractName, protocol.CreateLotEnabledSettingPath, trueValue)
}

func proposeSettingAuctionIsBidOnLotEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.AuctionSettingsContractName, protocol.BidOnLotEnabledSettingPath, trueValue)
}

func proposeSettingAuctionLotMinimumEthValue(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.AuctionSettingsContractName, protocol.LotMinimumEthValueSettingPath, trueValue)
}

func proposeSettingAuctionLotMaximumEthValue(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.AuctionSettingsContractName, protocol.LotMaximumEthValueSettingPath, trueValue)
}

func proposeSettingAuctionLotDuration(c *cli.Context, value time.Duration) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(c, protocol.AuctionSettingsContractName, protocol.LotDurationSettingPath, trueValue)
}

func proposeSettingAuctionLotStartingPriceRatio(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.AuctionSettingsContractName, protocol.LotStartingPriceRatioSettingPath, trueValue)
}

func proposeSettingAuctionLotReservePriceRatio(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.AuctionSettingsContractName, protocol.LotReservePriceRatioSettingPath, trueValue)
}

func proposeSettingDepositIsDepositingEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.DepositSettingsContractName, protocol.DepositEnabledSettingPath, trueValue)
}

func proposeSettingDepositAreDepositAssignmentsEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.DepositSettingsContractName, protocol.AssignDepositsEnabledSettingPath, trueValue)
}

func proposeSettingDepositMinimumDeposit(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.DepositSettingsContractName, protocol.MinimumDepositSettingPath, trueValue)
}

func proposeSettingDepositMaximumDepositPoolSize(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.DepositSettingsContractName, protocol.MaximumDepositPoolSizeSettingPath, trueValue)
}

func proposeSettingDepositMaximumAssignmentsPerDeposit(c *cli.Context, value uint64) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.DepositSettingsContractName, protocol.MaximumDepositAssignmentsSettingPath, trueValue)
}

func proposeSettingDepositMaximumSocialisedAssignmentsPerDeposit(c *cli.Context, value uint64) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.DepositSettingsContractName, protocol.MaximumSocializedDepositAssignmentsSettingPath, trueValue)
}

func proposeSettingDepositExpressQueueRate(c *cli.Context, value uint64) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.DepositSettingsContractName, protocol.ExpressQueueRatePath, trueValue)
}

func proposeSettingDepositExpressQueueTicketsBaseProvision(c *cli.Context, value uint64) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.DepositSettingsContractName, protocol.ExpressQueueTicketsBaseProvisionPath, trueValue)
}

func proposeSettingDepositDepositFee(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.DepositSettingsContractName, protocol.DepositFeeSettingPath, trueValue)
}

func proposeSettingMinipoolIsSubmitWithdrawableEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.MinipoolSettingsContractName, protocol.MinipoolSubmitWithdrawableEnabledSettingPath, trueValue)
}

func proposeSettingMinipoolLaunchTimeout(c *cli.Context, value time.Duration) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(c, protocol.MinipoolSettingsContractName, protocol.MinipoolLaunchTimeoutSettingPath, trueValue)
}

func proposeSettingMinipoolIsBondReductionEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.MinipoolSettingsContractName, protocol.BondReductionEnabledSettingPath, trueValue)
}

func proposeSettingMinipoolMaximumCount(c *cli.Context, value uint64) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.MinipoolSettingsContractName, protocol.MaximumMinipoolCountSettingPath, trueValue)
}

func proposeSettingMinipoolUserDistributeWindowStart(c *cli.Context, value time.Duration) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(c, protocol.MinipoolSettingsContractName, protocol.MinipoolUserDistributeWindowStartSettingPath, trueValue)
}

func proposeSettingMinipoolUserDistributeWindowLength(c *cli.Context, value time.Duration) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(c, protocol.MinipoolSettingsContractName, protocol.MinipoolUserDistributeWindowLengthSettingPath, trueValue)
}

func proposeSettingNetworkOracleDaoConsensusThreshold(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.NetworkSettingsContractName, protocol.NodeConsensusThresholdSettingPath, trueValue)
}

func proposeSettingNetworkNodePenaltyThreshold(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.NetworkSettingsContractName, protocol.NetworkPenaltyThresholdSettingPath, trueValue)
}

func proposeSettingNetworkPerPenaltyRate(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.NetworkSettingsContractName, protocol.NetworkPenaltyPerRateSettingPath, trueValue)
}

func proposeSettingNetworkIsSubmitBalancesEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.NetworkSettingsContractName, protocol.SubmitBalancesEnabledSettingPath, trueValue)
}

func proposeSettingNetworkSubmitBalancesFrequency(c *cli.Context, value time.Duration) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(c, protocol.NetworkSettingsContractName, protocol.SubmitBalancesFrequencySettingPath, trueValue)
}

func proposeSettingNetworkIsSubmitPricesEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.NetworkSettingsContractName, protocol.SubmitPricesEnabledSettingPath, trueValue)
}

func proposeSettingNetworkSubmitPricesFrequency(c *cli.Context, value time.Duration) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(c, protocol.NetworkSettingsContractName, protocol.SubmitPricesFrequencySettingPath, trueValue)
}

func proposeSettingNetworkMinimumNodeFee(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.NetworkSettingsContractName, protocol.MinimumNodeFeeSettingPath, trueValue)
}

func proposeSettingNetworkTargetNodeFee(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.NetworkSettingsContractName, protocol.TargetNodeFeeSettingPath, trueValue)
}

func proposeSettingNetworkMaximumNodeFee(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.NetworkSettingsContractName, protocol.MaximumNodeFeeSettingPath, trueValue)
}

func proposeSettingNetworkNodeFeeDemandRange(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.NetworkSettingsContractName, protocol.NodeFeeDemandRangeSettingPath, trueValue)
}

func proposeSettingNetworkTargetRethCollateralRate(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.NetworkSettingsContractName, protocol.TargetRethCollateralRateSettingPath, trueValue)
}

func proposeSettingNetworkIsSubmitRewardsEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.NetworkSettingsContractName, protocol.SubmitRewardsEnabledSettingPath, trueValue)
}

func proposeSettingNodeIsRegistrationEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.NodeSettingsContractName, protocol.NodeRegistrationEnabledSettingPath, trueValue)
}

func proposeSettingNodeIsSmoothingPoolRegistrationEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.NodeSettingsContractName, protocol.SmoothingPoolRegistrationEnabledSettingPath, trueValue)
}

func proposeSettingNodeIsDepositingEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.NodeSettingsContractName, protocol.NodeDepositEnabledSettingPath, trueValue)
}

func proposeSettingNodeAreVacantMinipoolsEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.NodeSettingsContractName, protocol.VacantMinipoolsEnabledSettingPath, trueValue)
}

func proposeSettingNodeMinimumPerMinipoolStake(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.NodeSettingsContractName, protocol.MinimumPerMinipoolStakeSettingPath, trueValue)
}

func proposeSettingNodeMaximumPerMinipoolStake(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.NodeSettingsContractName, protocol.MaximumPerMinipoolStakeSettingPath, trueValue)
}

func proposeSettingProposalsVotePhase1Time(c *cli.Context, value time.Duration) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(c, protocol.ProposalsSettingsContractName, protocol.VotePhase1TimeSettingPath, trueValue)
}

func proposeSettingProposalsVotePhase2Time(c *cli.Context, value time.Duration) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(c, protocol.ProposalsSettingsContractName, protocol.VotePhase2TimeSettingPath, trueValue)
}

func proposeSettingProposalsVoteDelayTime(c *cli.Context, value time.Duration) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(c, protocol.ProposalsSettingsContractName, protocol.VoteDelayTimeSettingPath, trueValue)
}

func proposeSettingProposalsExecuteTime(c *cli.Context, value time.Duration) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(c, protocol.ProposalsSettingsContractName, protocol.ExecuteTimeSettingPath, trueValue)
}

func proposeSettingProposalsProposalBond(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.ProposalsSettingsContractName, protocol.ProposalBondSettingPath, trueValue)
}

func proposeSettingProposalsChallengeBond(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.ProposalsSettingsContractName, protocol.ChallengeBondSettingPath, trueValue)
}

func proposeSettingProposalsChallengePeriod(c *cli.Context, value time.Duration) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(c, protocol.ProposalsSettingsContractName, protocol.ChallengePeriodSettingPath, trueValue)
}

func proposeSettingProposalsQuorum(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.ProposalsSettingsContractName, protocol.ProposalQuorumSettingPath, trueValue)
}

func proposeSettingProposalsVetoQuorum(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.ProposalsSettingsContractName, protocol.ProposalVetoQuorumSettingPath, trueValue)
}

func proposeSettingProposalsMaxBlockAge(c *cli.Context, value uint64) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.ProposalsSettingsContractName, protocol.ProposalMaxBlockAgeSettingPath, trueValue)
}

func proposeSettingRewardsIntervalPeriods(c *cli.Context, value uint64) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.RewardsSettingsContractName, protocol.RewardsClaimIntervalPeriodsSettingPath, trueValue)
}

func proposeSettingSecurityMembersQuorum(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.SecuritySettingsContractName, protocol.SecurityMembersQuorumSettingPath, trueValue)
}

func proposeSettingSecurityMembersLeaveTime(c *cli.Context, value time.Duration) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(c, protocol.SecuritySettingsContractName, protocol.SecurityMembersLeaveTimeSettingPath, trueValue)
}

func proposeSettingSecurityProposalVoteTime(c *cli.Context, value time.Duration) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(c, protocol.SecuritySettingsContractName, protocol.SecurityProposalVoteTimeSettingPath, trueValue)
}

func proposeSettingSecurityProposalExecuteTime(c *cli.Context, value time.Duration) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(c, protocol.SecuritySettingsContractName, protocol.SecurityProposalExecuteTimeSettingPath, trueValue)
}

func proposeSettingSecurityProposalActionTime(c *cli.Context, value time.Duration) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(c, protocol.SecuritySettingsContractName, protocol.SecurityProposalActionTimeSettingPath, trueValue)
}

func proposeSettingNetworkAllowListedControllers(c *cli.Context, value []common.Address) error {
	strs := make([]string, len(value))
	for i, addr := range value {
		strs[i] = addr.Hex()
	}
	trueValue := strings.Join(strs, "")
	return proposeSetting(c, protocol.NetworkSettingsContractName, protocol.NetworkAllowListedControllersPath, trueValue)
}

// Master general proposal function
func proposeSetting(c *cli.Context, contract string, setting string, value string) error {
	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check if proposal can be made
	canPropose, err := rp.PDAOCanProposeSetting(contract, setting, value)
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
		if canPropose.IsRplLockingDisallowed {
			fmt.Println("Please enable RPL locking using the command 'rocketpool node allow-rpl-locking' to raise proposals.")
		}
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canPropose.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm("Are you sure you want to submit this proposal?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Submit proposal
	response, err := rp.PDAOProposeSetting(contract, setting, value, canPropose.BlockNumber)
	if err != nil {
		return err
	}

	fmt.Printf("Submitting proposal...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully submitted a %s setting update proposal.\n", setting)
	return nil
}
