package pdao

import (
	"fmt"
	"math/big"
	"time"

	"github.com/rocket-pool/smartnode/bindings/settings/protocol"

	cliutils "github.com/rocket-pool/smartnode/rocketpool-cli/cli"
	"github.com/rocket-pool/smartnode/shared/math"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"

	"github.com/rocket-pool/smartnode/rocketpool-cli/cli/prompt"
)

func proposeSettingAuctionIsCreateLotEnabled(value bool, yes bool, toJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.AuctionSettingsContractName, protocol.CreateLotEnabledSettingPath, trueValue, yes, toJson)
}

func proposeSettingAuctionIsBidOnLotEnabled(value bool, yes bool, toJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.AuctionSettingsContractName, protocol.BidOnLotEnabledSettingPath, trueValue, yes, toJson)
}

func proposeSettingAuctionLotMinimumEthValue(value *big.Int, yes bool, toJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.AuctionSettingsContractName, protocol.LotMinimumEthValueSettingPath, trueValue, yes, toJson)
}

func proposeSettingAuctionLotMaximumEthValue(value *big.Int, yes bool, toJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.AuctionSettingsContractName, protocol.LotMaximumEthValueSettingPath, trueValue, yes, toJson)
}

func proposeSettingAuctionLotDuration(value time.Duration, yes bool, toJson string) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.AuctionSettingsContractName, protocol.LotDurationSettingPath, trueValue, yes, toJson)
}

func proposeSettingAuctionLotStartingPriceRatio(value *big.Int, yes bool, toJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.AuctionSettingsContractName, protocol.LotStartingPriceRatioSettingPath, trueValue, yes, toJson)
}

func proposeSettingAuctionLotReservePriceRatio(value *big.Int, yes bool, toJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.AuctionSettingsContractName, protocol.LotReservePriceRatioSettingPath, trueValue, yes, toJson)
}

func proposeSettingDepositIsDepositingEnabled(value bool, yes bool, toJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.DepositSettingsContractName, protocol.DepositEnabledSettingPath, trueValue, yes, toJson)
}

func proposeSettingDepositAreDepositAssignmentsEnabled(value bool, yes bool, toJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.DepositSettingsContractName, protocol.AssignDepositsEnabledSettingPath, trueValue, yes, toJson)
}

func proposeSettingDepositMinimumDeposit(value *big.Int, yes bool, toJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.DepositSettingsContractName, protocol.MinimumDepositSettingPath, trueValue, yes, toJson)
}

func proposeSettingDepositMaximumDepositPoolSize(value *big.Int, yes bool, toJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.DepositSettingsContractName, protocol.MaximumDepositPoolSizeSettingPath, trueValue, yes, toJson)
}

func proposeSettingDepositMaximumAssignmentsPerDeposit(value uint64, yes bool, toJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.DepositSettingsContractName, protocol.MaximumDepositAssignmentsSettingPath, trueValue, yes, toJson)
}

func proposeSettingDepositMaximumSocialisedAssignmentsPerDeposit(value uint64, yes bool, toJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.DepositSettingsContractName, protocol.MaximumSocializedDepositAssignmentsSettingPath, trueValue, yes, toJson)
}

func proposeSettingDepositExpressQueueRate(value uint64, yes bool, toJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.DepositSettingsContractName, protocol.ExpressQueueRatePath, trueValue, yes, toJson)
}

func proposeSettingDepositExpressQueueTicketsBaseProvision(value uint64, yes bool, toJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.DepositSettingsContractName, protocol.ExpressQueueTicketsBaseProvisionPath, trueValue, yes, toJson)
}

func proposeSettingDepositDepositFee(value *big.Int, yes bool, toJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.DepositSettingsContractName, protocol.DepositFeeSettingPath, trueValue, yes, toJson)
}

func proposeSettingMinipoolIsSubmitWithdrawableEnabled(value bool, yes bool, toJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.MinipoolSettingsContractName, protocol.MinipoolSubmitWithdrawableEnabledSettingPath, trueValue, yes, toJson)
}

func proposeSettingMinipoolLaunchTimeout(value time.Duration, yes bool, toJson string) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.MinipoolSettingsContractName, protocol.MinipoolLaunchTimeoutSettingPath, trueValue, yes, toJson)
}

func proposeSettingMinipoolIsBondReductionEnabled(value bool, yes bool, toJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.MinipoolSettingsContractName, protocol.BondReductionEnabledSettingPath, trueValue, yes, toJson)
}

func proposeSettingMinipoolMaximumCount(value uint64, yes bool, toJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.MinipoolSettingsContractName, protocol.MaximumMinipoolCountSettingPath, trueValue, yes, toJson)
}

func proposeSettingMinipoolUserDistributeWindowStart(value time.Duration, yes bool, toJson string) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.MinipoolSettingsContractName, protocol.MinipoolUserDistributeWindowStartSettingPath, trueValue, yes, toJson)
}

func proposeSettingMinipoolUserDistributeWindowLength(value time.Duration, yes bool, toJson string) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.MinipoolSettingsContractName, protocol.MinipoolUserDistributeWindowLengthSettingPath, trueValue, yes, toJson)
}

func proposeSettingNetworkOracleDaoConsensusThreshold(value *big.Int, yes bool, toJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.NodeConsensusThresholdSettingPath, trueValue, yes, toJson)
}

func proposeSettingNetworkNodePenaltyThreshold(value *big.Int, yes bool, toJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.NetworkPenaltyThresholdSettingPath, trueValue, yes, toJson)
}

func proposeSettingNetworkPerPenaltyRate(value *big.Int, yes bool, toJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.NetworkPenaltyPerRateSettingPath, trueValue, yes, toJson)
}

func proposeSettingNetworkIsSubmitBalancesEnabled(value bool, yes bool, toJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.SubmitBalancesEnabledSettingPath, trueValue, yes, toJson)
}

func proposeSettingNetworkSubmitBalancesFrequency(value time.Duration, yes bool, toJson string) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.SubmitBalancesFrequencySettingPath, trueValue, yes, toJson)
}

func proposeSettingNetworkIsSubmitPricesEnabled(value bool, yes bool, toJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.SubmitPricesEnabledSettingPath, trueValue, yes, toJson)
}

func proposeSettingNetworkSubmitPricesFrequency(value time.Duration, yes bool, toJson string) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.SubmitPricesFrequencySettingPath, trueValue, yes, toJson)
}

func proposeSettingNetworkMinimumNodeFee(value *big.Int, yes bool, toJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.MinimumNodeFeeSettingPath, trueValue, yes, toJson)
}

func proposeSettingNetworkTargetNodeFee(value *big.Int, yes bool, toJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.TargetNodeFeeSettingPath, trueValue, yes, toJson)
}

func proposeSettingNetworkMaximumNodeFee(value *big.Int, yes bool, toJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.MaximumNodeFeeSettingPath, trueValue, yes, toJson)
}

func proposeSettingNetworkNodeFeeDemandRange(value *big.Int, yes bool, toJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.NodeFeeDemandRangeSettingPath, trueValue, yes, toJson)
}

func proposeSettingNetworkTargetRethCollateralRate(value *big.Int, yes bool, toJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.TargetRethCollateralRateSettingPath, trueValue, yes, toJson)
}

func proposeSettingNetworkIsSubmitRewardsEnabled(value bool, yes bool, toJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.SubmitRewardsEnabledSettingPath, trueValue, yes, toJson)
}

func proposeSettingNodeIsRegistrationEnabled(value bool, yes bool, toJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.NodeSettingsContractName, protocol.NodeRegistrationEnabledSettingPath, trueValue, yes, toJson)
}

func proposeSettingNodeIsSmoothingPoolRegistrationEnabled(value bool, yes bool, toJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.NodeSettingsContractName, protocol.SmoothingPoolRegistrationEnabledSettingPath, trueValue, yes, toJson)
}

func proposeSettingNodeIsDepositingEnabled(value bool, yes bool, toJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.NodeSettingsContractName, protocol.NodeDepositEnabledSettingPath, trueValue, yes, toJson)
}

func proposeSettingNodeAreVacantMinipoolsEnabled(value bool, yes bool, toJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.NodeSettingsContractName, protocol.VacantMinipoolsEnabledSettingPath, trueValue, yes, toJson)
}

func proposeSettingNodeMinimumLegacyRplStake(value *big.Int, yes bool, toJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.NodeSettingsContractName, protocol.MinimumLegacyRplStakePath, trueValue, yes, toJson)
}

func proposeSettingReducedBond(value *big.Int, yes bool, toJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.NodeSettingsContractName, protocol.ReducedBondSettingPath, trueValue, yes, toJson)
}

func proposeSettingNodeUnstakingPeriod(value time.Duration, yes bool, toJson string) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.NodeSettingsContractName, protocol.NodeUnstakingPeriodSettingPath, trueValue, yes, toJson)
}

func proposeSettingProposalsVotePhase1Time(value time.Duration, yes bool, toJson string) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.ProposalsSettingsContractName, protocol.VotePhase1TimeSettingPath, trueValue, yes, toJson)
}

func proposeSettingProposalsVotePhase2Time(value time.Duration, yes bool, toJson string) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.ProposalsSettingsContractName, protocol.VotePhase2TimeSettingPath, trueValue, yes, toJson)
}

func proposeSettingProposalsVoteDelayTime(value time.Duration, yes bool, toJson string) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.ProposalsSettingsContractName, protocol.VoteDelayTimeSettingPath, trueValue, yes, toJson)
}

func proposeSettingProposalsExecuteTime(value time.Duration, yes bool, toJson string) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.ProposalsSettingsContractName, protocol.ExecuteTimeSettingPath, trueValue, yes, toJson)
}

func proposeSettingProposalsProposalBond(value *big.Int, yes bool, toJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.ProposalsSettingsContractName, protocol.ProposalBondSettingPath, trueValue, yes, toJson)
}

func proposeSettingProposalsChallengeBond(value *big.Int, yes bool, toJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.ProposalsSettingsContractName, protocol.ChallengeBondSettingPath, trueValue, yes, toJson)
}

func proposeSettingProposalsChallengePeriod(value time.Duration, yes bool, toJson string) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.ProposalsSettingsContractName, protocol.ChallengePeriodSettingPath, trueValue, yes, toJson)
}

func proposeSettingProposalsQuorum(value *big.Int, yes bool, toJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.ProposalsSettingsContractName, protocol.ProposalQuorumSettingPath, trueValue, yes, toJson)
}

func proposeSettingProposalsVetoQuorum(value *big.Int, yes bool, toJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.ProposalsSettingsContractName, protocol.ProposalVetoQuorumSettingPath, trueValue, yes, toJson)
}

func proposeSettingProposalsMaxBlockAge(value uint64, yes bool, toJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.ProposalsSettingsContractName, protocol.ProposalMaxBlockAgeSettingPath, trueValue, yes, toJson)
}

func proposeSettingRewardsIntervalPeriods(value uint64, yes bool, toJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.RewardsSettingsContractName, protocol.RewardsClaimIntervalPeriodsSettingPath, trueValue, yes, toJson)
}

func proposeSettingSecurityMembersQuorum(value *big.Int, yes bool, toJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.SecuritySettingsContractName, protocol.SecurityMembersQuorumSettingPath, trueValue, yes, toJson)
}

func proposeSettingSecurityMembersLeaveTime(value time.Duration, yes bool, toJson string) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.SecuritySettingsContractName, protocol.SecurityMembersLeaveTimeSettingPath, trueValue, yes, toJson)
}

func proposeSettingSecurityProposalVoteTime(value time.Duration, yes bool, toJson string) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.SecuritySettingsContractName, protocol.SecurityProposalVoteTimeSettingPath, trueValue, yes, toJson)
}

func proposeSettingSecurityProposalExecuteTime(value time.Duration, yes bool, toJson string) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.SecuritySettingsContractName, protocol.SecurityProposalExecuteTimeSettingPath, trueValue, yes, toJson)
}

func proposeSettingSecurityProposalActionTime(value time.Duration, yes bool, toJson string) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.SecuritySettingsContractName, protocol.SecurityProposalActionTimeSettingPath, trueValue, yes, toJson)
}

func proposeSettingMegapoolTimeBeforeDissolve(value time.Duration, yes bool, toJson string) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(protocol.MegapoolSettingsContractName, protocol.MegapoolTimeBeforeDissolveSettingsPath, trueValue, yes, toJson)
}

func proposeSettingMaximumMegapoolEthPenalty(value *big.Int, yes bool, toJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.MegapoolSettingsContractName, protocol.MegapoolMaximumMegapoolEthPenaltyPath, trueValue, yes, toJson)
}

func proposeSettingMegapoolNotifyThreshold(value uint64, yes bool, toJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.MegapoolSettingsContractName, protocol.MegapoolNotifyThresholdPath, trueValue, yes, toJson)
}

func proposeSettingMegapoolLateNotifyFine(value *big.Int, yes bool, toJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.MegapoolSettingsContractName, protocol.MegapoolLateNotifyFinePath, trueValue, yes, toJson)
}

func proposeSettingMegapoolDissolvePenalty(value *big.Int, yes bool, toJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.MegapoolSettingsContractName, protocol.MegapoolDissolvePenaltyPath, trueValue, yes, toJson)
}

func proposeSettingMegapoolUserDistributeDelay(value uint64, yes bool, toJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.MegapoolSettingsContractName, protocol.MegapoolUserDistributeDelayPath, trueValue, yes, toJson)
}

func proposeSettingMegapoolUserDistributeDelayWithShortfall(value uint64, yes bool, toJson string) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.MegapoolSettingsContractName, protocol.MegapoolUserDistributeDelayShortfallPath, trueValue, yes, toJson)
}

func proposeSettingPenaltyThreshold(value *big.Int, yes bool, toJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.MegapoolSettingsContractName, protocol.MegapoolPenaltyThreshold, trueValue, yes, toJson)
}

func proposeSettingNodeCommissionShare(value *big.Int, yes bool, toJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.NetworkNodeCommissionSharePath, trueValue, yes, toJson)
}

func proposeSettingNodeCommissionShareSecurityCouncilAdder(value *big.Int, yes bool, toJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.NetworkNodeCommissionShareSecurityCouncilAdderPath, trueValue, yes, toJson)
}

func proposeSettingVoterShare(value *big.Int, yes bool, toJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.NetworkVoterSharePath, trueValue, yes, toJson)
}

func proposeSettingPDAOShare(value *big.Int, yes bool, toJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.NetworkPDAOSharePath, trueValue, yes, toJson)
}

func proposeMaxNodeShareSecurityCouncilAdder(value *big.Int, yes bool, toJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.NetworkMaxNodeShareSecurityCouncilAdderPath, trueValue, yes, toJson)
}

func proposeMaxRethBalanceDelta(value *big.Int, yes bool, toJson string) error {
	trueValue := value.String()
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.NetworkMaxRethBalanceDeltaPath, trueValue, yes, toJson)
}

// Master general proposal function
func proposeSetting(contract string, setting string, value string, yes bool, toJson string) error {
	if toJson != "" {
		return writeSettingToBatchJSON(toJson, contract, setting, value)
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
