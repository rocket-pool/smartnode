package pdao

import (
	"fmt"
	"math/big"
	"time"

	"github.com/rocket-pool/smartnode/bindings/settings/protocol"

	protocol131 "github.com/rocket-pool/smartnode/bindings/legacy/v1.3.1/protocol"
	cliutils "github.com/rocket-pool/smartnode/rocketpool-cli/cli"
	"github.com/rocket-pool/smartnode/shared/math"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"

	"github.com/rocket-pool/smartnode/rocketpool-cli/cli/prompt"
)

func proposeSettingAuctionIsCreateLotEnabled(value bool, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.AuctionSettingsContractName, protocol.CreateLotEnabledSettingPath, trueValue, yes, generateJson)
}

func proposeSettingAuctionIsBidOnLotEnabled(value bool, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.AuctionSettingsContractName, protocol.BidOnLotEnabledSettingPath, trueValue, yes, generateJson)
}

func proposeSettingAuctionLotMinimumEthValue(value *big.Int, yes bool, generateJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.AuctionSettingsContractName, protocol.LotMinimumEthValueSettingPath, trueValue, yes, generateJson)
}

func proposeSettingAuctionLotMaximumEthValue(value *big.Int, yes bool, generateJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.AuctionSettingsContractName, protocol.LotMaximumEthValueSettingPath, trueValue, yes, generateJson)
}

func proposeSettingAuctionLotDuration(value time.Duration, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.AuctionSettingsContractName, protocol.LotDurationSettingPath, trueValue, yes, generateJson)
}

func proposeSettingAuctionLotStartingPriceRatio(value *big.Int, yes bool, generateJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.AuctionSettingsContractName, protocol.LotStartingPriceRatioSettingPath, trueValue, yes, generateJson)
}

func proposeSettingAuctionLotReservePriceRatio(value *big.Int, yes bool, generateJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.AuctionSettingsContractName, protocol.LotReservePriceRatioSettingPath, trueValue, yes, generateJson)
}

func proposeSettingDepositIsDepositingEnabled(value bool, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.DepositSettingsContractName, protocol.DepositEnabledSettingPath, trueValue, yes, generateJson)
}

func proposeSettingDepositAreDepositAssignmentsEnabled(value bool, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.DepositSettingsContractName, protocol.AssignDepositsEnabledSettingPath, trueValue, yes, generateJson)
}

func proposeSettingDepositMinimumDeposit(value *big.Int, yes bool, generateJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.DepositSettingsContractName, protocol.MinimumDepositSettingPath, trueValue, yes, generateJson)
}

func proposeSettingDepositMaximumDepositPoolSize(value *big.Int, yes bool, generateJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.DepositSettingsContractName, protocol.MaximumDepositPoolSizeSettingPath, trueValue, yes, generateJson)
}

func proposeSettingDepositMaximumAssignmentsPerDeposit(value uint64, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.DepositSettingsContractName, protocol.MaximumDepositAssignmentsSettingPath, trueValue, yes, generateJson)
}

func proposeSettingDepositMaximumSocialisedAssignmentsPerDeposit(value uint64, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.DepositSettingsContractName, protocol.MaximumSocializedDepositAssignmentsSettingPath, trueValue, yes, generateJson)
}

func proposeSettingDepositExpressQueueRate(value uint64, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.DepositSettingsContractName, protocol.ExpressQueueRatePath, trueValue, yes, generateJson)
}

func proposeSettingDepositExpressQueueTicketsBaseProvision(value uint64, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.DepositSettingsContractName, protocol.ExpressQueueTicketsBaseProvisionPath, trueValue, yes, generateJson)
}

func proposeSettingDepositDepositFee(value *big.Int, yes bool, generateJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.DepositSettingsContractName, protocol.DepositFeeSettingPath, trueValue, yes, generateJson)
}

func proposeSettingMinipoolIsSubmitWithdrawableEnabled(value bool, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.MinipoolSettingsContractName, protocol.MinipoolSubmitWithdrawableEnabledSettingPath, trueValue, yes, generateJson)
}

func proposeSettingMinipoolLaunchTimeout(value time.Duration, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.MinipoolSettingsContractName, protocol.MinipoolLaunchTimeoutSettingPath, trueValue, yes, generateJson)
}

func proposeSettingMinipoolIsBondReductionEnabled(value bool, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.MinipoolSettingsContractName, protocol.BondReductionEnabledSettingPath, trueValue, yes, generateJson)
}

func proposeSettingMinipoolMaximumCount(value uint64, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.MinipoolSettingsContractName, protocol.MaximumMinipoolCountSettingPath, trueValue, yes, generateJson)
}

func proposeSettingMinipoolUserDistributeWindowStart(value time.Duration, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.MinipoolSettingsContractName, protocol.MinipoolUserDistributeWindowStartSettingPath, trueValue, yes, generateJson)
}

func proposeSettingMinipoolUserDistributeWindowLength(value time.Duration, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.MinipoolSettingsContractName, protocol.MinipoolUserDistributeWindowLengthSettingPath, trueValue, yes, generateJson)
}

func proposeSettingNetworkOracleDaoConsensusThreshold(value *big.Int, yes bool, generateJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.NodeConsensusThresholdSettingPath, trueValue, yes, generateJson)
}

func proposeSettingNetworkNodePenaltyThreshold(value *big.Int, yes bool, generateJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.NetworkPenaltyThresholdSettingPath, trueValue, yes, generateJson)
}

func proposeSettingNetworkPerPenaltyRate(value *big.Int, yes bool, generateJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.NetworkPenaltyPerRateSettingPath, trueValue, yes, generateJson)
}

func proposeSettingNetworkIsSubmitBalancesEnabled(value bool, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.SubmitBalancesEnabledSettingPath, trueValue, yes, generateJson)
}

func proposeSettingNetworkSubmitBalancesFrequency(value time.Duration, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.SubmitBalancesFrequencySettingPath, trueValue, yes, generateJson)
}

func proposeSettingNetworkIsSubmitPricesEnabled(value bool, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.SubmitPricesEnabledSettingPath, trueValue, yes, generateJson)
}

func proposeSettingNetworkSubmitPricesFrequency(value time.Duration, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.SubmitPricesFrequencySettingPath, trueValue, yes, generateJson)
}

func proposeSettingNetworkMinimumNodeFee(value *big.Int, yes bool, generateJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.MinimumNodeFeeSettingPath, trueValue, yes, generateJson)
}

func proposeSettingNetworkTargetNodeFee(value *big.Int, yes bool, generateJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.TargetNodeFeeSettingPath, trueValue, yes, generateJson)
}

func proposeSettingNetworkMaximumNodeFee(value *big.Int, yes bool, generateJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.MaximumNodeFeeSettingPath, trueValue, yes, generateJson)
}

func proposeSettingNetworkNodeFeeDemandRange(value *big.Int, yes bool, generateJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.NodeFeeDemandRangeSettingPath, trueValue, yes, generateJson)
}

func proposeSettingNetworkTargetRethCollateralRate(value *big.Int, yes bool, generateJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.TargetRethCollateralRateSettingPath, trueValue, yes, generateJson)
}

func proposeSettingNetworkIsSubmitRewardsEnabled(value bool, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.SubmitRewardsEnabledSettingPath, trueValue, yes, generateJson)
}

func proposeSettingNodeIsRegistrationEnabled(value bool, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.NodeSettingsContractName, protocol.NodeRegistrationEnabledSettingPath, trueValue, yes, generateJson)
}

func proposeSettingNodeIsSmoothingPoolRegistrationEnabled(value bool, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.NodeSettingsContractName, protocol.SmoothingPoolRegistrationEnabledSettingPath, trueValue, yes, generateJson)
}

func proposeSettingNodeIsDepositingEnabled(value bool, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.NodeSettingsContractName, protocol.NodeDepositEnabledSettingPath, trueValue, yes, generateJson)
}

func proposeSettingNodeAreVacantMinipoolsEnabled(value bool, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.NodeSettingsContractName, protocol.VacantMinipoolsEnabledSettingPath, trueValue, yes, generateJson)
}

func proposeSettingNodeMinimumPerMinipoolStake(value *big.Int, yes bool, generateJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.NodeSettingsContractName, protocol131.MinimumPerMinipoolStakeSettingPath, trueValue, yes, generateJson)
}

func proposeSettingNodeMaximumPerMinipoolStake(value *big.Int, yes bool, generateJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.NodeSettingsContractName, protocol131.MaximumPerMinipoolStakeSettingPath, trueValue, yes, generateJson)
}

func proposeSettingNodeMinimumLegacyRplStake(value *big.Int, yes bool, generateJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.NodeSettingsContractName, protocol.MinimumLegacyRplStakePath, trueValue, yes, generateJson)
}

func proposeSettingReducedBond(value *big.Int, yes bool, generateJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.NodeSettingsContractName, protocol.ReducedBondSettingPath, trueValue, yes, generateJson)
}

func proposeSettingNodeUnstakingPeriod(value time.Duration, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.NodeSettingsContractName, protocol.NodeUnstakingPeriodSettingPath, trueValue, yes, generateJson)
}

func proposeSettingProposalsVotePhase1Time(value time.Duration, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.ProposalsSettingsContractName, protocol.VotePhase1TimeSettingPath, trueValue, yes, generateJson)
}

func proposeSettingProposalsVotePhase2Time(value time.Duration, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.ProposalsSettingsContractName, protocol.VotePhase2TimeSettingPath, trueValue, yes, generateJson)
}

func proposeSettingProposalsVoteDelayTime(value time.Duration, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.ProposalsSettingsContractName, protocol.VoteDelayTimeSettingPath, trueValue, yes, generateJson)
}

func proposeSettingProposalsExecuteTime(value time.Duration, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.ProposalsSettingsContractName, protocol.ExecuteTimeSettingPath, trueValue, yes, generateJson)
}

func proposeSettingProposalsProposalBond(value *big.Int, yes bool, generateJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.ProposalsSettingsContractName, protocol.ProposalBondSettingPath, trueValue, yes, generateJson)
}

func proposeSettingProposalsChallengeBond(value *big.Int, yes bool, generateJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.ProposalsSettingsContractName, protocol.ChallengeBondSettingPath, trueValue, yes, generateJson)
}

func proposeSettingProposalsChallengePeriod(value time.Duration, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.ProposalsSettingsContractName, protocol.ChallengePeriodSettingPath, trueValue, yes, generateJson)
}

func proposeSettingProposalsQuorum(value *big.Int, yes bool, generateJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.ProposalsSettingsContractName, protocol.ProposalQuorumSettingPath, trueValue, yes, generateJson)
}

func proposeSettingProposalsVetoQuorum(value *big.Int, yes bool, generateJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.ProposalsSettingsContractName, protocol.ProposalVetoQuorumSettingPath, trueValue, yes, generateJson)
}

func proposeSettingProposalsMaxBlockAge(value uint64, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.ProposalsSettingsContractName, protocol.ProposalMaxBlockAgeSettingPath, trueValue, yes, generateJson)
}

func proposeSettingRewardsIntervalPeriods(value uint64, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.RewardsSettingsContractName, protocol.RewardsClaimIntervalPeriodsSettingPath, trueValue, yes, generateJson)
}

func proposeSettingSecurityMembersQuorum(value *big.Int, yes bool, generateJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.SecuritySettingsContractName, protocol.SecurityMembersQuorumSettingPath, trueValue, yes, generateJson)
}

func proposeSettingSecurityMembersLeaveTime(value time.Duration, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.SecuritySettingsContractName, protocol.SecurityMembersLeaveTimeSettingPath, trueValue, yes, generateJson)
}

func proposeSettingSecurityProposalVoteTime(value time.Duration, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.SecuritySettingsContractName, protocol.SecurityProposalVoteTimeSettingPath, trueValue, yes, generateJson)
}

func proposeSettingSecurityProposalExecuteTime(value time.Duration, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.SecuritySettingsContractName, protocol.SecurityProposalExecuteTimeSettingPath, trueValue, yes, generateJson)
}

func proposeSettingSecurityProposalActionTime(value time.Duration, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.SecuritySettingsContractName, protocol.SecurityProposalActionTimeSettingPath, trueValue, yes, generateJson)
}

func proposeSettingMegapoolTimeBeforeDissolve(value time.Duration, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.MegapoolSettingsContractName, protocol.MegapoolTimeBeforeDissolveSettingsPath, trueValue, yes, generateJson)
}

func proposeSettingMaximumMegapoolEthPenalty(value *big.Int, yes bool, generateJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.MegapoolSettingsContractName, protocol.MegapoolMaximumMegapoolEthPenaltyPath, trueValue, yes, generateJson)
}

func proposeSettingMegapoolNotifyThreshold(value uint64, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.MegapoolSettingsContractName, protocol.MegapoolNotifyThresholdPath, trueValue, yes, generateJson)
}

func proposeSettingMegapoolLateNotifyFine(value *big.Int, yes bool, generateJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.MegapoolSettingsContractName, protocol.MegapoolLateNotifyFinePath, trueValue, yes, generateJson)
}

func proposeSettingMegapoolDissolvePenalty(value *big.Int, yes bool, generateJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.MegapoolSettingsContractName, protocol.MegapoolDissolvePenaltyPath, trueValue, yes, generateJson)
}

func proposeSettingMegapoolUserDistributeDelay(value uint64, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.MegapoolSettingsContractName, protocol.MegapoolUserDistributeDelayPath, trueValue, yes, generateJson)
}

func proposeSettingMegapoolUserDistributeDelayWithShortfall(value uint64, yes bool, generateJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.MegapoolSettingsContractName, protocol.MegapoolUserDistributeDelayShortfallPath, trueValue, yes, generateJson)
}

func proposeSettingPenaltyThreshold(value *big.Int, yes bool, generateJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.MegapoolSettingsContractName, protocol.MegapoolPenaltyThreshold, trueValue, yes, generateJson)
}

func proposeSettingNodeCommissionShare(value *big.Int, yes bool, generateJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.NetworkNodeCommissionSharePath, trueValue, yes, generateJson)
}

func proposeSettingNodeCommissionShareSecurityCouncilAdder(value *big.Int, yes bool, generateJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.NetworkNodeCommissionShareSecurityCouncilAdderPath, trueValue, yes, generateJson)
}

func proposeSettingVoterShare(value *big.Int, yes bool, generateJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.NetworkVoterSharePath, trueValue, yes, generateJson)
}

func proposeSettingPDAOShare(value *big.Int, yes bool, generateJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.NetworkPDAOSharePath, trueValue, yes, generateJson)
}

func proposeMaxNodeShareSecurityCouncilAdder(value *big.Int, yes bool, generateJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.NetworkMaxNodeShareSecurityCouncilAdderPath, trueValue, yes, generateJson)
}

func proposeMaxRethBalanceDelta(value *big.Int, yes bool, generateJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.NetworkMaxRethBalanceDeltaPath, trueValue, yes, generateJson)
}

// Master general proposal function
func proposeSetting(contract string, setting string, value string, yes bool, generateJson string) error {
	if protocol.IsHoustonOnlySetting(setting) {
		fmt.Println("This command no longer available in Saturn.")
		return nil
	}

	if generateJson != "" {
		return writeSettingToBatchJSON(generateJson, contract, setting, value)
	}

	// Get RP client
	rp, err := rocketpool.NewClient().WithReady()
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
				math.WeiToEth(big.NewInt(0).Sub(canPropose.StakedRpl, canPropose.LockedRpl)), math.WeiToEth(canPropose.ProposalBond),
			)
		}
		if canPropose.IsRplLockingDisallowed {
			fmt.Println("Please enable RPL locking using the command 'rocketpool node allow-rpl-locking' to raise proposals.")
		}
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canPropose.GasLimits, rp, yes)
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if prompt.Declined(yes, "Are you sure you want to submit this proposal?") {
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
