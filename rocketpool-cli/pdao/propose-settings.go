package pdao

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"

	protocol131 "github.com/rocket-pool/smartnode/bindings/legacy/v1.3.1/protocol"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"

	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

func proposeSettingAuctionIsCreateLotEnabled(value bool, yes bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.AuctionSettingsContractName, protocol.CreateLotEnabledSettingPath, trueValue, yes)
}

func proposeSettingAuctionIsBidOnLotEnabled(value bool, yes bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.AuctionSettingsContractName, protocol.BidOnLotEnabledSettingPath, trueValue, yes)
}

func proposeSettingAuctionLotMinimumEthValue(value *big.Int, yes bool) error {
	trueValue := value.String()
	return proposeSetting(protocol.AuctionSettingsContractName, protocol.LotMinimumEthValueSettingPath, trueValue, yes)
}

func proposeSettingAuctionLotMaximumEthValue(value *big.Int, yes bool) error {
	trueValue := value.String()
	return proposeSetting(protocol.AuctionSettingsContractName, protocol.LotMaximumEthValueSettingPath, trueValue, yes)
}

func proposeSettingAuctionLotDuration(value time.Duration, yes bool) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.AuctionSettingsContractName, protocol.LotDurationSettingPath, trueValue, yes)
}

func proposeSettingAuctionLotStartingPriceRatio(value *big.Int, yes bool) error {
	trueValue := value.String()
	return proposeSetting(protocol.AuctionSettingsContractName, protocol.LotStartingPriceRatioSettingPath, trueValue, yes)
}

func proposeSettingAuctionLotReservePriceRatio(value *big.Int, yes bool) error {
	trueValue := value.String()
	return proposeSetting(protocol.AuctionSettingsContractName, protocol.LotReservePriceRatioSettingPath, trueValue, yes)
}

func proposeSettingDepositIsDepositingEnabled(value bool, yes bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.DepositSettingsContractName, protocol.DepositEnabledSettingPath, trueValue, yes)
}

func proposeSettingDepositAreDepositAssignmentsEnabled(value bool, yes bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.DepositSettingsContractName, protocol.AssignDepositsEnabledSettingPath, trueValue, yes)
}

func proposeSettingDepositMinimumDeposit(value *big.Int, yes bool) error {
	trueValue := value.String()
	return proposeSetting(protocol.DepositSettingsContractName, protocol.MinimumDepositSettingPath, trueValue, yes)
}

func proposeSettingDepositMaximumDepositPoolSize(value *big.Int, yes bool) error {
	trueValue := value.String()
	return proposeSetting(protocol.DepositSettingsContractName, protocol.MaximumDepositPoolSizeSettingPath, trueValue, yes)
}

func proposeSettingDepositMaximumAssignmentsPerDeposit(value uint64, yes bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.DepositSettingsContractName, protocol.MaximumDepositAssignmentsSettingPath, trueValue, yes)
}

func proposeSettingDepositMaximumSocialisedAssignmentsPerDeposit(value uint64, yes bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.DepositSettingsContractName, protocol.MaximumSocializedDepositAssignmentsSettingPath, trueValue, yes)
}

func proposeSettingDepositExpressQueueRate(value uint64, yes bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.DepositSettingsContractName, protocol.ExpressQueueRatePath, trueValue, yes)
}

func proposeSettingDepositExpressQueueTicketsBaseProvision(value uint64, yes bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.DepositSettingsContractName, protocol.ExpressQueueTicketsBaseProvisionPath, trueValue, yes)
}

func proposeSettingDepositDepositFee(value *big.Int, yes bool) error {
	trueValue := value.String()
	return proposeSetting(protocol.DepositSettingsContractName, protocol.DepositFeeSettingPath, trueValue, yes)
}

func proposeSettingMinipoolIsSubmitWithdrawableEnabled(value bool, yes bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.MinipoolSettingsContractName, protocol.MinipoolSubmitWithdrawableEnabledSettingPath, trueValue, yes)
}

func proposeSettingMinipoolLaunchTimeout(value time.Duration, yes bool) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.MinipoolSettingsContractName, protocol.MinipoolLaunchTimeoutSettingPath, trueValue, yes)
}

func proposeSettingMinipoolIsBondReductionEnabled(value bool, yes bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.MinipoolSettingsContractName, protocol.BondReductionEnabledSettingPath, trueValue, yes)
}

func proposeSettingMinipoolMaximumCount(value uint64, yes bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.MinipoolSettingsContractName, protocol.MaximumMinipoolCountSettingPath, trueValue, yes)
}

func proposeSettingMinipoolUserDistributeWindowStart(value time.Duration, yes bool) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.MinipoolSettingsContractName, protocol.MinipoolUserDistributeWindowStartSettingPath, trueValue, yes)
}

func proposeSettingMinipoolUserDistributeWindowLength(value time.Duration, yes bool) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.MinipoolSettingsContractName, protocol.MinipoolUserDistributeWindowLengthSettingPath, trueValue, yes)
}

func proposeSettingNetworkOracleDaoConsensusThreshold(value *big.Int, yes bool) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.NodeConsensusThresholdSettingPath, trueValue, yes)
}

func proposeSettingNetworkNodePenaltyThreshold(value *big.Int, yes bool) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.NetworkPenaltyThresholdSettingPath, trueValue, yes)
}

func proposeSettingNetworkPerPenaltyRate(value *big.Int, yes bool) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.NetworkPenaltyPerRateSettingPath, trueValue, yes)
}

func proposeSettingNetworkIsSubmitBalancesEnabled(value bool, yes bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.SubmitBalancesEnabledSettingPath, trueValue, yes)
}

func proposeSettingNetworkSubmitBalancesFrequency(value time.Duration, yes bool) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.SubmitBalancesFrequencySettingPath, trueValue, yes)
}

func proposeSettingNetworkIsSubmitPricesEnabled(value bool, yes bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.SubmitPricesEnabledSettingPath, trueValue, yes)
}

func proposeSettingNetworkSubmitPricesFrequency(value time.Duration, yes bool) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.SubmitPricesFrequencySettingPath, trueValue, yes)
}

func proposeSettingNetworkMinimumNodeFee(value *big.Int, yes bool) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.MinimumNodeFeeSettingPath, trueValue, yes)
}

func proposeSettingNetworkTargetNodeFee(value *big.Int, yes bool) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.TargetNodeFeeSettingPath, trueValue, yes)
}

func proposeSettingNetworkMaximumNodeFee(value *big.Int, yes bool) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.MaximumNodeFeeSettingPath, trueValue, yes)
}

func proposeSettingNetworkNodeFeeDemandRange(value *big.Int, yes bool) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.NodeFeeDemandRangeSettingPath, trueValue, yes)
}

func proposeSettingNetworkTargetRethCollateralRate(value *big.Int, yes bool) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.TargetRethCollateralRateSettingPath, trueValue, yes)
}

func proposeSettingNetworkIsSubmitRewardsEnabled(value bool, yes bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.SubmitRewardsEnabledSettingPath, trueValue, yes)
}

func proposeSettingNodeIsRegistrationEnabled(value bool, yes bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.NodeSettingsContractName, protocol.NodeRegistrationEnabledSettingPath, trueValue, yes)
}

func proposeSettingNodeIsSmoothingPoolRegistrationEnabled(value bool, yes bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.NodeSettingsContractName, protocol.SmoothingPoolRegistrationEnabledSettingPath, trueValue, yes)
}

func proposeSettingNodeIsDepositingEnabled(value bool, yes bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.NodeSettingsContractName, protocol.NodeDepositEnabledSettingPath, trueValue, yes)
}

func proposeSettingNodeAreVacantMinipoolsEnabled(value bool, yes bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.NodeSettingsContractName, protocol.VacantMinipoolsEnabledSettingPath, trueValue, yes)
}

func proposeSettingNodeMinimumPerMinipoolStake(value *big.Int, yes bool) error {
	trueValue := value.String()
	return proposeSetting(protocol.NodeSettingsContractName, protocol131.MinimumPerMinipoolStakeSettingPath, trueValue, yes)
}

func proposeSettingNodeMaximumPerMinipoolStake(value *big.Int, yes bool) error {
	trueValue := value.String()
	return proposeSetting(protocol.NodeSettingsContractName, protocol131.MaximumPerMinipoolStakeSettingPath, trueValue, yes)
}

func proposeSettingNodeMinimumLegacyRplStake(value *big.Int, yes bool) error {
	trueValue := value.String()
	return proposeSetting(protocol.NodeSettingsContractName, protocol.MinimumLegacyRplStakePath, trueValue, yes)
}

func proposeSettingReducedBond(value *big.Int, yes bool) error {
	trueValue := value.String()
	return proposeSetting(protocol.NodeSettingsContractName, protocol.ReducedBondSettingPath, trueValue, yes)
}

func proposeSettingNodeUnstakingPeriod(value time.Duration, yes bool) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.NodeSettingsContractName, protocol.NodeUnstakingPeriodSettingPath, trueValue, yes)
}

func proposeSettingProposalsVotePhase1Time(value time.Duration, yes bool) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.ProposalsSettingsContractName, protocol.VotePhase1TimeSettingPath, trueValue, yes)
}

func proposeSettingProposalsVotePhase2Time(value time.Duration, yes bool) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.ProposalsSettingsContractName, protocol.VotePhase2TimeSettingPath, trueValue, yes)
}

func proposeSettingProposalsVoteDelayTime(value time.Duration, yes bool) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.ProposalsSettingsContractName, protocol.VoteDelayTimeSettingPath, trueValue, yes)
}

func proposeSettingProposalsExecuteTime(value time.Duration, yes bool) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.ProposalsSettingsContractName, protocol.ExecuteTimeSettingPath, trueValue, yes)
}

func proposeSettingProposalsProposalBond(value *big.Int, yes bool) error {
	trueValue := value.String()
	return proposeSetting(protocol.ProposalsSettingsContractName, protocol.ProposalBondSettingPath, trueValue, yes)
}

func proposeSettingProposalsChallengeBond(value *big.Int, yes bool) error {
	trueValue := value.String()
	return proposeSetting(protocol.ProposalsSettingsContractName, protocol.ChallengeBondSettingPath, trueValue, yes)
}

func proposeSettingProposalsChallengePeriod(value time.Duration, yes bool) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.ProposalsSettingsContractName, protocol.ChallengePeriodSettingPath, trueValue, yes)
}

func proposeSettingProposalsQuorum(value *big.Int, yes bool) error {
	trueValue := value.String()
	return proposeSetting(protocol.ProposalsSettingsContractName, protocol.ProposalQuorumSettingPath, trueValue, yes)
}

func proposeSettingProposalsVetoQuorum(value *big.Int, yes bool) error {
	trueValue := value.String()
	return proposeSetting(protocol.ProposalsSettingsContractName, protocol.ProposalVetoQuorumSettingPath, trueValue, yes)
}

func proposeSettingProposalsMaxBlockAge(value uint64, yes bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.ProposalsSettingsContractName, protocol.ProposalMaxBlockAgeSettingPath, trueValue, yes)
}

func proposeSettingRewardsIntervalPeriods(value uint64, yes bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.RewardsSettingsContractName, protocol.RewardsClaimIntervalPeriodsSettingPath, trueValue, yes)
}

func proposeSettingSecurityMembersQuorum(value *big.Int, yes bool) error {
	trueValue := value.String()
	return proposeSetting(protocol.SecuritySettingsContractName, protocol.SecurityMembersQuorumSettingPath, trueValue, yes)
}

func proposeSettingSecurityMembersLeaveTime(value time.Duration, yes bool) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.SecuritySettingsContractName, protocol.SecurityMembersLeaveTimeSettingPath, trueValue, yes)
}

func proposeSettingSecurityProposalVoteTime(value time.Duration, yes bool) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.SecuritySettingsContractName, protocol.SecurityProposalVoteTimeSettingPath, trueValue, yes)
}

func proposeSettingSecurityProposalExecuteTime(value time.Duration, yes bool) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.SecuritySettingsContractName, protocol.SecurityProposalExecuteTimeSettingPath, trueValue, yes)
}

func proposeSettingSecurityProposalActionTime(value time.Duration, yes bool) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.SecuritySettingsContractName, protocol.SecurityProposalActionTimeSettingPath, trueValue, yes)
}

func proposeSettingNetworkAllowListedControllers(value []common.Address, yes bool) error {
	strs := make([]string, len(value))
	for i, addr := range value {
		strs[i] = addr.Hex()
	}
	trueValue := strings.Join(strs, "")
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.NetworkAllowListedControllersPath, trueValue, yes)
}

func proposeSettingMegapoolTimeBeforeDissolve(value time.Duration, yes bool) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.MegapoolSettingsContractName, protocol.MegapoolTimeBeforeDissolveSettingsPath, trueValue, yes)
}

func proposeSettingMaximumMegapoolEthPenalty(value *big.Int, yes bool) error {
	trueValue := value.String()
	return proposeSetting(protocol.MegapoolSettingsContractName, protocol.MegapoolMaximumMegapoolEthPenaltyPath, trueValue, yes)
}

func proposeSettingMegapoolNotifyThreshold(value uint64, yes bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.MegapoolSettingsContractName, protocol.MegapoolNotifyThresholdPath, trueValue, yes)
}

func proposeSettingMegapoolLateNotifyFine(value *big.Int, yes bool) error {
	trueValue := value.String()
	return proposeSetting(protocol.MegapoolSettingsContractName, protocol.MegapoolLateNotifyFinePath, trueValue, yes)
}

func proposeSettingMegapoolDissolvePenalty(value *big.Int, yes bool) error {
	trueValue := value.String()
	return proposeSetting(protocol.MegapoolSettingsContractName, protocol.MegapoolDissolvePenaltyPath, trueValue, yes)
}

func proposeSettingMegapoolUserDistributeDelay(value uint64, yes bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.MegapoolSettingsContractName, protocol.MegapoolUserDistributeDelayPath, trueValue, yes)
}

func proposeSettingMegapoolUserDistributeDelayWithShortfall(value uint64, yes bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.MegapoolSettingsContractName, protocol.MegapoolUserDistributeDelayShortfallPath, trueValue, yes)
}

func proposeSettingPenaltyThreshold(value *big.Int, yes bool) error {
	trueValue := value.String()
	return proposeSetting(protocol.MegapoolSettingsContractName, protocol.MegapoolPenaltyThreshold, trueValue, yes)
}

func proposeSettingNodeCommissionShare(value *big.Int, yes bool) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.NetworkNodeCommissionSharePath, trueValue, yes)
}

func proposeSettingNodeCommissionShareSecurityCouncilAdder(value *big.Int, yes bool) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.NetworkNodeCommissionShareSecurityCouncilAdderPath, trueValue, yes)
}

func proposeSettingVoterShare(value *big.Int, yes bool) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.NetworkVoterSharePath, trueValue, yes)
}

func proposeSettingPDAOShare(value *big.Int, yes bool) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.NetworkPDAOSharePath, trueValue, yes)
}

func proposeMaxNodeShareSecurityCouncilAdder(value *big.Int, yes bool) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.NetworkMaxNodeShareSecurityCouncilAdderPath, trueValue, yes)
}

func proposeMaxRethBalanceDelta(value *big.Int, yes bool) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.NetworkMaxRethBalanceDeltaPath, trueValue, yes)
}

// Master general proposal function
func proposeSetting(contract string, setting string, value string, yes bool) error {
	// Get RP client
	rp, err := rocketpool.NewClient().WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	if isHoustonOnlySetting(setting) {
		fmt.Println("This command no longer available in Saturn.")
		return nil
	}

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
	err = gas.AssignMaxFeeAndLimit(canPropose.GasInfo, rp, yes)
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(yes || prompt.Confirm("Are you sure you want to submit this proposal?")) {
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

// Returns true if the given setting is only available on Houston 1.3.1 (before the Saturn upgrade).
func isHoustonOnlySetting(setting string) bool {

	// Map of Houston only settings
	houstonOnlySettings := map[string]struct{}{
		protocol131.MinimumPerMinipoolStakeSettingPath: {},
		protocol131.MaximumPerMinipoolStakeSettingPath: {},
	}

	_, exists := houstonOnlySettings[setting]
	return exists
}
