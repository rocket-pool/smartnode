package protocol

import (
	"fmt"

	"github.com/rocket-pool/smartnode/bindings/types"
)

// Proposal setting type names written to --to-json files.
const (
	ProposalSettingTypeNameUint256 = "uint256"
	ProposalSettingTypeNameBool    = "bool"
	ProposalSettingTypeNameAddress = "address"
)

var (
	ErrUnknownPDAOSetting      = fmt.Errorf("unknown protocol DAO setting")
	ErrUnsupportedBatchSetting = fmt.Errorf("setting type is not supported in a multi-setting proposal")
	ErrDuplicateBatchSetting   = fmt.Errorf("duplicate setting in multi-setting proposal")
	ErrEmptyBatchSettings      = fmt.Errorf("multi-setting proposal must contain at least one setting")
)

// settingKind is the on-chain value type of a protocol DAO setting.
type settingKind uint8

const (
	settingKindUint256 settingKind = iota
	settingKindBool
	settingKindAddress
	settingKindAddressList
)

// pdaoSettingKinds maps contract name -> setting path -> value type for every
// setting the Smart Node can propose.
var pdaoSettingKinds = map[string]map[string]settingKind{
	AuctionSettingsContractName: {
		CreateLotEnabledSettingPath:      settingKindBool,
		BidOnLotEnabledSettingPath:       settingKindBool,
		LotMinimumEthValueSettingPath:    settingKindUint256,
		LotMaximumEthValueSettingPath:    settingKindUint256,
		LotDurationSettingPath:           settingKindUint256,
		LotStartingPriceRatioSettingPath: settingKindUint256,
		LotReservePriceRatioSettingPath:  settingKindUint256,
	},
	DepositSettingsContractName: {
		DepositEnabledSettingPath:                      settingKindBool,
		AssignDepositsEnabledSettingPath:               settingKindBool,
		MinimumDepositSettingPath:                      settingKindUint256,
		MaximumDepositPoolSizeSettingPath:              settingKindUint256,
		MaximumDepositAssignmentsSettingPath:           settingKindUint256,
		MaximumSocializedDepositAssignmentsSettingPath: settingKindUint256,
		DepositFeeSettingPath:                          settingKindUint256,
		ExpressQueueRatePath:                           settingKindUint256,
		ExpressQueueTicketsBaseProvisionPath:           settingKindUint256,
	},
	MinipoolSettingsContractName: {
		MinipoolSubmitWithdrawableEnabledSettingPath:  settingKindBool,
		MinipoolLaunchTimeoutSettingPath:              settingKindUint256,
		BondReductionEnabledSettingPath:               settingKindBool,
		MaximumMinipoolCountSettingPath:               settingKindUint256,
		MinipoolUserDistributeWindowStartSettingPath:  settingKindUint256,
		MinipoolUserDistributeWindowLengthSettingPath: settingKindUint256,
	},
	NetworkSettingsContractName: {
		NodeConsensusThresholdSettingPath:                  settingKindUint256,
		SubmitBalancesEnabledSettingPath:                   settingKindBool,
		SubmitBalancesFrequencySettingPath:                 settingKindUint256,
		SubmitPricesEnabledSettingPath:                     settingKindBool,
		SubmitPricesFrequencySettingPath:                   settingKindUint256,
		MinimumNodeFeeSettingPath:                          settingKindUint256,
		TargetNodeFeeSettingPath:                           settingKindUint256,
		MaximumNodeFeeSettingPath:                          settingKindUint256,
		NodeFeeDemandRangeSettingPath:                      settingKindUint256,
		TargetRethCollateralRateSettingPath:                settingKindUint256,
		NetworkPenaltyThresholdSettingPath:                 settingKindUint256,
		NetworkPenaltyPerRateSettingPath:                   settingKindUint256,
		SubmitRewardsEnabledSettingPath:                    settingKindBool,
		NetworkAllowListedControllersPath:                  settingKindAddressList,
		NetworkNodeCommissionSharePath:                     settingKindUint256,
		NetworkNodeCommissionShareSecurityCouncilAdderPath: settingKindUint256,
		NetworkVoterSharePath:                              settingKindUint256,
		NetworkPDAOSharePath:                               settingKindUint256,
		NetworkMaxNodeShareSecurityCouncilAdderPath:        settingKindUint256,
		NetworkMaxRethBalanceDeltaPath:                     settingKindUint256,
	},
	NodeSettingsContractName: {
		NodeRegistrationEnabledSettingPath:          settingKindBool,
		SmoothingPoolRegistrationEnabledSettingPath: settingKindBool,
		NodeDepositEnabledSettingPath:               settingKindBool,
		VacantMinipoolsEnabledSettingPath:           settingKindBool,
		MinimumLegacyRplStakePath:                   settingKindUint256,
		ReducedBondSettingPath:                      settingKindUint256,
		NodeUnstakingPeriodSettingPath:              settingKindUint256,
	},
	ProposalsSettingsContractName: {
		VotePhase1TimeSettingPath:      settingKindUint256,
		VotePhase2TimeSettingPath:      settingKindUint256,
		VoteDelayTimeSettingPath:       settingKindUint256,
		ExecuteTimeSettingPath:         settingKindUint256,
		ProposalBondSettingPath:        settingKindUint256,
		ChallengeBondSettingPath:       settingKindUint256,
		ChallengePeriodSettingPath:     settingKindUint256,
		ProposalQuorumSettingPath:      settingKindUint256,
		ProposalVetoQuorumSettingPath:  settingKindUint256,
		ProposalMaxBlockAgeSettingPath: settingKindUint256,
	},
	RewardsSettingsContractName: {
		RewardsClaimIntervalPeriodsSettingPath: settingKindUint256,
	},
	SecuritySettingsContractName: {
		SecurityMembersQuorumSettingPath:       settingKindUint256,
		SecurityMembersLeaveTimeSettingPath:    settingKindUint256,
		SecurityProposalVoteTimeSettingPath:    settingKindUint256,
		SecurityProposalExecuteTimeSettingPath: settingKindUint256,
		SecurityProposalActionTimeSettingPath:  settingKindUint256,
	},
	MegapoolSettingsContractName: {
		MegapoolTimeBeforeDissolveSettingsPath:   settingKindUint256,
		MegapoolMaximumMegapoolEthPenaltyPath:    settingKindUint256,
		MegapoolNotifyThresholdPath:              settingKindUint256,
		MegapoolLateNotifyFinePath:               settingKindUint256,
		MegapoolDissolvePenaltyPath:              settingKindUint256,
		MegapoolUserDistributeDelayPath:          settingKindUint256,
		MegapoolUserDistributeDelayShortfallPath: settingKindUint256,
		MegapoolPenaltyThreshold:                 settingKindUint256,
	},
}

// GetProposalSettingType returns the on-chain type used by proposalSettingMulti
// for the given contract/setting pair.
func GetProposalSettingType(contract string, setting string) (types.ProposalSettingType, error) {
	kinds, ok := pdaoSettingKinds[contract]
	if !ok {
		return 0, fmt.Errorf("%w: [%s - %s]", ErrUnknownPDAOSetting, contract, setting)
	}
	kind, ok := kinds[setting]
	if !ok {
		return 0, fmt.Errorf("%w: [%s - %s]", ErrUnknownPDAOSetting, contract, setting)
	}
	switch kind {
	case settingKindUint256:
		return types.ProposalSettingType_Uint256, nil
	case settingKindBool:
		return types.ProposalSettingType_Bool, nil
	case settingKindAddress:
		return types.ProposalSettingType_Address, nil
	case settingKindAddressList:
		return 0, fmt.Errorf("%w: %s (address lists cannot be batched)", ErrUnsupportedBatchSetting, setting)
	default:
		return 0, fmt.Errorf("%w: [%s - %s]", ErrUnknownPDAOSetting, contract, setting)
	}
}

// ProposalSettingTypeName returns the JSON type string for a setting type.
func ProposalSettingTypeName(settingType types.ProposalSettingType) string {
	switch settingType {
	case types.ProposalSettingType_Uint256:
		return ProposalSettingTypeNameUint256
	case types.ProposalSettingType_Bool:
		return ProposalSettingTypeNameBool
	case types.ProposalSettingType_Address:
		return ProposalSettingTypeNameAddress
	default:
		return ""
	}
}

// ParseProposalSettingTypeName converts a JSON type string to a setting type.
func ParseProposalSettingTypeName(name string) (types.ProposalSettingType, error) {
	switch name {
	case ProposalSettingTypeNameUint256, "uint":
		return types.ProposalSettingType_Uint256, nil
	case ProposalSettingTypeNameBool:
		return types.ProposalSettingType_Bool, nil
	case ProposalSettingTypeNameAddress:
		return types.ProposalSettingType_Address, nil
	default:
		return 0, fmt.Errorf("unknown setting type %q", name)
	}
}
