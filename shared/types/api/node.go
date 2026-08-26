package api

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/bindings/tokens"
	"github.com/rocket-pool/smartnode/bindings/transactions/gaslimit"
	rptypes "github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/rocketpool-cli/cli/color"
	"github.com/rocket-pool/smartnode/rocketpool/feerecipient"
	"github.com/rocket-pool/smartnode/shared/services/rewards"
)

type NodeStatusResponse struct {
	APIResponse
	Warning                                  string          `json:"warning"`
	AccountAddress                           common.Address  `json:"accountAddress"`
	AccountAddressFormatted                  string          `json:"accountAddressFormatted"`
	PrimaryWithdrawalAddress                 common.Address  `json:"primaryWithdrawalAddress"`
	PrimaryWithdrawalAddressFormatted        string          `json:"primaryWithdrawalAddressFormatted"`
	PendingPrimaryWithdrawalAddress          common.Address  `json:"pendingPrimaryWithdrawalAddress"`
	PendingPrimaryWithdrawalAddressFormatted string          `json:"pendingPrimaryWithdrawalAddressFormatted"`
	IsRPLWithdrawalAddressSet                bool            `json:"isRPLWithdrawalAddressSet"`
	RPLWithdrawalAddress                     common.Address  `json:"rplWithdrawalAddress"`
	RPLWithdrawalAddressFormatted            string          `json:"rplWithdrawalAddressFormatted"`
	PendingRPLWithdrawalAddress              common.Address  `json:"pendingRPLWithdrawalAddress"`
	PendingRPLWithdrawalAddressFormatted     string          `json:"pendingRPLWithdrawalAddressFormatted"`
	IsRPLLockingAllowed                      bool            `json:"isRPLLockingAllowed"`
	NodeRPLLocked                            *big.Int        `json:"nodeRPLLocked"`
	Registered                               bool            `json:"registered"`
	Trusted                                  bool            `json:"trusted"`
	TimezoneLocation                         string          `json:"timezoneLocation"`
	AccountBalances                          tokens.Balances `json:"accountBalances"`
	PrimaryWithdrawalBalances                tokens.Balances `json:"primaryWithdrawalBalances"`
	RPLWithdrawalBalances                    tokens.Balances `json:"rplWithdrawalBalances"`
	TotalRplStake                            *big.Int        `json:"totalRplStake"`
	RplStakeMegapool                         *big.Int        `json:"rplStakeMegapool"`
	RplStakeLegacy                           *big.Int        `json:"rplStakeLegacy"`
	RplStakeThreshold                        *big.Int        `json:"rplStakeThreshold"`
	RplStakeThresholdFraction                float64         `json:"rplStakeThresholdFraction"`
	BorrowedCollateralRatio                  float64         `json:"borrowedCollateralRatio"`
	BondedCollateralRatio                    float64         `json:"bondedCollateralRatio"`
	PendingMinimumRplStake                   *big.Int        `json:"pendingMinimumRplStake"`
	PendingMaximumRplStake                   *big.Int        `json:"pendingMaximumRplStake"`
	PendingBorrowedCollateralRatio           float64         `json:"pendingBorrowedCollateralRatio"`
	PendingBondedCollateralRatio             float64         `json:"pendingBondedCollateralRatio"`
	OnchainVotingDelegate                    common.Address  `json:"onchainVotingDelegate"`
	OnchainVotingDelegateFormatted           string          `json:"onchainVotingDelegateFormatted"`
	MinipoolLimit                            uint64          `json:"minipoolLimit"`
	EthBorrowed                              *big.Int        `json:"ethBorrowed"`
	EthBorrowedLimit                         *big.Int        `json:"ethBorrowedLimit"`
	PendingBorrowAmount                      *big.Int        `json:"pendingBorrowAmount"`
	CreditBalance                            *big.Int        `json:"creditBalance"`
	CreditAndEthOnBehalfBalance              *big.Int        `json:"creditAndEthOnBehalfBalance"`
	EthOnBehalfBalance                       *big.Int        `json:"ethOnBehalfBalance"`
	UsableCreditAndEthOnBehalfBalance        *big.Int        `json:"usableCreditAndEthOnBehalfBalance"`
	MinipoolCounts                           struct {
		Total               int `json:"total"`
		Initialized         int `json:"initialized"`
		Prelaunch           int `json:"prelaunch"`
		Staking             int `json:"staking"`
		Withdrawable        int `json:"withdrawable"`
		Dissolved           int `json:"dissolved"`
		RefundAvailable     int `json:"refundAvailable"`
		WithdrawalAvailable int `json:"withdrawalAvailable"`
		CloseAvailable      int `json:"closeAvailable"`
		Finalised           int `json:"finalised"`
	} `json:"minipoolCounts"`
	IsFeeDistributorInitialized bool                      `json:"isFeeDistributorInitialized"`
	FeeRecipientInfo            feerecipient.Details      `json:"feeRecipientInfo"`
	FeeDistributorBalance       *big.Int                  `json:"feeDistributorBalance"`
	PenalizedMinipools          map[common.Address]uint64 `json:"penalizedMinipools"`
	SnapshotResponse            struct {
		Error                   string                 `json:"error"`
		ProposalVotes           []SnapshotProposalVote `json:"proposalVotes"`
		ActiveSnapshotProposals []SnapshotProposal     `json:"activeSnapshotProposals"`
	} `json:"snapshotResponse"`
	SignallingAddress            common.Address    `json:"signallingAddress"`
	SignallingAddressFormatted   string            `json:"signallingAddressFormatted"`
	Minipools                    []MinipoolDetails `json:"minipools"`
	LatestDelegate               common.Address    `json:"latestDelegate"`
	MegapoolDeployed             bool              `json:"megapoolDeployed"`
	MegapoolAddress              common.Address    `json:"megapoolAddress"`
	MegapoolActiveValidatorCount uint16            `json:"megapoolActiveValidatorCount"`
	MegapoolNodeDebt             *big.Int          `json:"megapoolNodeDebt"`
	MegapoolRefundValue          *big.Int          `json:"megapoolRefundValue"`
	ExpressTicketCount           uint64            `json:"expressTicketCount"`
	ExpressTicketsProvisioned    bool              `json:"expressTicketsProvisioned"`
	UnstakingRPL                 *big.Int          `json:"unstakingRPL"`
	LastRPLUnstakeTime           time.Time         `json:"lastRPLUnstakeTime"`
	UnstakingPeriodDuration      time.Duration     `json:"unstakingPeriodDuration"`
	LatestBlockTime              time.Time         `json:"latestBlockTime"`
	UnclaimedRewards             *big.Int          `json:"unclaimedRewards"`
	ReducedBond                  *big.Int          `json:"reducedBond"`
}

type NodeAlert struct {
	// Enum: [unprocessed active suppressed]
	State string `json:"state"`
	// NOTE: Alertmanager puts "description" and "summary" in annotations and "alertname" is in labels (along with any configured labels and annotations).
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

func (n NodeAlert) IsActive() bool {
	return n.State == "active"
}

func (n NodeAlert) IsSuppressed() bool {
	return n.State == "suppressed"
}

func (n NodeAlert) Name() string {
	value, ok := n.Labels["alertname"]
	if !ok {
		return ""
	}
	return value
}

func (n NodeAlert) Summary() string {
	value, ok := n.Annotations["summary"]
	if !ok {
		return ""
	}
	return value
}

func (n NodeAlert) Description() string {
	value, ok := n.Annotations["description"]
	if !ok {
		return ""
	}
	return value
}

func (n NodeAlert) Severity() string {
	value, ok := n.Labels["severity"]
	if !ok {
		return ""
	}
	return value
}

func (n NodeAlert) ColorString() string {
	suppressed := ""
	if n.IsSuppressed() {
		suppressed = " (suppressed)"
	}
	alertColor := color.Yellow
	if n.Severity() == "critical" {
		alertColor = color.Red
	}
	header := alertColor(fmt.Sprintf("%s%s", n.Severity(), suppressed))
	return fmt.Sprintf("%s %s: %s", header, n.Summary(), n.Description())
}

type CanRegisterNodeResponse struct {
	APIResponse
	CanRegister          bool            `json:"canRegister"`
	AlreadyRegistered    bool            `json:"alreadyRegistered"`
	RegistrationDisabled bool            `json:"registrationDisabled"`
	GasLimits            gaslimit.Limits `json:"gasLimits"`
}
type RegisterNodeResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type CanProvisionExpressTicketsResponse struct {
	APIResponse
	CanProvision       bool            `json:"canProvision"`
	AlreadyProvisioned bool            `json:"alreadyProvisioned"`
	GasLimits          gaslimit.Limits `json:"gasLimits"`
}
type ProvisionExpressTicketsResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type CanSetNodePrimaryWithdrawalAddressResponse struct {
	APIResponse
	CanSet    bool            `json:"canSet"`
	GasLimits gaslimit.Limits `json:"gasLimits"`
}
type SetNodePrimaryWithdrawalAddressResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type CanConfirmNodePrimaryWithdrawalAddressResponse struct {
	APIResponse
	CanConfirm bool            `json:"canConfirm"`
	GasLimits  gaslimit.Limits `json:"gasLimits"`
}
type ConfirmNodePrimaryWithdrawalAddressResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type CanSetNodeRPLWithdrawalAddressResponse struct {
	APIResponse
	CanSet                bool            `json:"canSet"`
	PrimaryAddressDiffers bool            `json:"primaryAddressDiffers"`
	RPLAddressDiffers     bool            `json:"rplAddressDiffers"`
	RPLStake              *big.Int        `json:"rplStake"`
	GasLimits             gaslimit.Limits `json:"gasLimits"`
}
type SetNodeRPLWithdrawalAddressResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type CanConfirmNodeRPLWithdrawalAddressResponse struct {
	APIResponse
	CanConfirm bool            `json:"canConfirm"`
	GasLimits  gaslimit.Limits `json:"gasLimits"`
}
type ConfirmNodeRPLWithdrawalAddressResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type GetNodePrimaryWithdrawalAddressResponse struct {
	APIResponse
	Address common.Address `json:"address"`
}

type GetNodePendingPrimaryWithdrawalAddressResponse struct {
	APIResponse
	Address common.Address `json:"address"`
}

type CanSetNodeTimezoneResponse struct {
	APIResponse
	CanSet    bool            `json:"canSet"`
	GasLimits gaslimit.Limits `json:"gasLimits"`
}
type SetNodeTimezoneResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type CanNodeSwapRplResponse struct {
	APIResponse
	CanSwap             bool            `json:"canSwap"`
	InsufficientBalance bool            `json:"insufficientBalance"`
	GasLimits           gaslimit.Limits `json:"GasLimits"`
}
type NodeSwapRplApproveGasResponse struct {
	APIResponse
	GasLimits gaslimit.Limits `json:"gasLimits"`
}
type NodeSwapRplApproveResponse struct {
	APIResponse
	ApproveTxHash common.Hash `json:"approveTxHash"`
}
type NodeSwapRplSwapResponse struct {
	APIResponse
	SwapTxHash common.Hash `json:"swapTxHash"`
}
type NodeSwapRplAllowanceResponse struct {
	APIResponse
	Allowance *big.Int `json:"allowance"`
}

type CanNodeStakeRplResponse struct {
	APIResponse
	CanStake            bool            `json:"canStake"`
	InsufficientBalance bool            `json:"insufficientBalance"`
	InConsensus         bool            `json:"inConsensus"`
	GasLimits           gaslimit.Limits `json:"gasLimits"`
}
type NodeStakeRplApproveGasResponse struct {
	APIResponse
	GasLimits gaslimit.Limits `json:"gasLimits"`
}
type NodeStakeRplApproveResponse struct {
	APIResponse
	ApproveTxHash common.Hash `json:"approveTxHash"`
}
type NodeStakeRplStakeResponse struct {
	APIResponse
	StakeTxHash common.Hash `json:"stakeTxHash"`
}
type NodeStakeRplAllowanceResponse struct {
	APIResponse
	Allowance *big.Int `json:"allowance"`
}

type CanSetRplLockingAllowedResponse struct {
	APIResponse
	CanSet    bool            `json:"canSet"`
	GasLimits gaslimit.Limits `json:"gasLimits"`
}

type SetRplLockingAllowedResponse struct {
	APIResponse
	SetTxHash common.Hash `json:"setTxHash"`
}
type CanSetStakeRplForAllowedResponse struct {
	APIResponse
	CanSet    bool            `json:"canSet"`
	GasLimits gaslimit.Limits `json:"gasLimits"`
}
type SetStakeRplForAllowedResponse struct {
	APIResponse
	SetTxHash common.Hash `json:"setTxHash"`
}
type CanNodeWithdrawEthResponse struct {
	APIResponse
	CanWithdraw                   bool            `json:"canWithdraw"`
	InsufficientBalance           bool            `json:"insufficientBalance"`
	HasDifferentWithdrawalAddress bool            `json:"hasDifferentWithdrawalAddress"`
	GasLimits                     gaslimit.Limits `json:"gasLimits"`
}
type NodeWithdrawEthResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}
type CanNodeWithdrawCreditResponse struct {
	APIResponse
	CanWithdraw         bool            `json:"canWithdraw"`
	InsufficientBalance bool            `json:"insufficientBalance"`
	GasLimits           gaslimit.Limits `json:"gasLimits"`
}
type NodeWithdrawCreditResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type CanNodeUnstakeRplResponse struct {
	APIResponse
	CanUnstake                       bool            `json:"canUnstake"`
	InsufficientBalance              bool            `json:"insufficientBalance"`
	HasDifferentRPLWithdrawalAddress bool            `json:"hasDifferentRPLWithdrawalAddress"`
	GasLimits                        gaslimit.Limits `json:"gasLimits"`
}
type NodeUnstakeRplResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}
type CanNodeUnstakeLegacyRplResponse struct {
	APIResponse
	CanUnstake                       bool            `json:"canUnstake"`
	InsufficientBalance              bool            `json:"insufficientBalance"`
	HasDifferentRPLWithdrawalAddress bool            `json:"hasDifferentRPLWithdrawalAddress"`
	BelowMaxRPLStake                 bool            `json:"belowMaxRPLStake"`
	GasLimits                        gaslimit.Limits `json:"gasLimits"`
}
type NodeUnstakeLegacyRplResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type NodeWithdrawRplResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}
type CanNodeWithdrawRplResponse struct {
	APIResponse
	CanWithdraw                      bool            `json:"canWithdraw"`
	InsufficientBalance              bool            `json:"insufficientBalance"`
	UnstakingPeriodActive            bool            `json:"unstakingPeriodActive"`
	HasDifferentRPLWithdrawalAddress bool            `json:"hasDifferentRPLWithdrawalAddress"`
	GasLimits                        gaslimit.Limits `json:"gasLimits"`
}
type CanNodeWithdrawRplv1_3_1Response struct {
	APIResponse
	CanWithdraw                      bool            `json:"canWithdraw"`
	InsufficientBalance              bool            `json:"insufficientBalance"`
	BelowMaxRPLStake                 bool            `json:"belowMaxRPLStake"`
	MinipoolsUndercollateralized     bool            `json:"minipoolsUndercollateralized"`
	WithdrawalDelayActive            bool            `json:"withdrawalDelayActive"`
	HasDifferentRPLWithdrawalAddress bool            `json:"hasDifferentRPLWithdrawalAddress"`
	GasLimits                        gaslimit.Limits `json:"gasLimits"`
}

type CanNodeDepositsResponse struct {
	APIResponse
	CanDeposit                       bool                      `json:"canDeposit"`
	CreditBalance                    *big.Int                  `json:"creditBalance"`
	UsableCreditBalance              *big.Int                  `json:"usableCreditBalance"`
	DepositBalance                   *big.Int                  `json:"depositBalance"`
	CanUseCredit                     bool                      `json:"canUseCredit"`
	NodeBalance                      *big.Int                  `json:"nodeBalance"`
	InsufficientBalance              bool                      `json:"insufficientBalance"`
	InsufficientBalanceWithoutCredit bool                      `json:"insufficientBalanceWithoutCredit"`
	InvalidAmount                    bool                      `json:"invalidAmount"`
	DepositDisabled                  bool                      `json:"depositDisabled"`
	InConsensus                      bool                      `json:"inConsensus"`
	NodeHasDebt                      bool                      `json:"nodeHasDebt"`
	MinipoolAddress                  common.Address            `json:"minipoolAddress"`
	MegapoolAddress                  common.Address            `json:"megapoolAddress"`
	ValidatorPubkeys                 []rptypes.ValidatorPubkey `json:"validatorPubkeys"`
	GasLimits                        gaslimit.Limits           `json:"gasLimits"`
}

type NodeDepositsResponse struct {
	APIResponse
	TxHash           common.Hash               `json:"txHash"`
	ValidatorPubkeys []rptypes.ValidatorPubkey `json:"validatorPubkeys"`
	ScrubPeriod      time.Duration             `json:"scrubPeriod"`
}

type CanCreateVacantMinipoolResponse struct {
	APIResponse
	CanDeposit           bool            `json:"canDeposit"`
	InsufficientRplStake bool            `json:"insufficientRplStake"`
	InvalidAmount        bool            `json:"invalidAmount"`
	DepositDisabled      bool            `json:"depositDisabled"`
	MinipoolAddress      common.Address  `json:"minipoolAddress"`
	GasLimits            gaslimit.Limits `json:"gasLimits"`
}
type CreateVacantMinipoolResponse struct {
	APIResponse
	TxHash                common.Hash    `json:"txHash"`
	MinipoolAddress       common.Address `json:"minipoolAddress"`
	ScrubPeriod           time.Duration  `json:"scrubPeriod"`
	WithdrawalCredentials common.Hash    `json:"withdrawalCredentials"`
}

type CanNodeSendResponse struct {
	APIResponse
	Balance             float64         `json:"balance"`
	TokenName           string          `json:"name"`
	TokenSymbol         string          `json:"symbol"`
	CanSend             bool            `json:"canSend"`
	InsufficientBalance bool            `json:"insufficientBalance"`
	GasLimits           gaslimit.Limits `json:"gasLimits"`
}
type NodeSendResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type CanNodeSendMessageResponse struct {
	APIResponse
	GasLimits gaslimit.Limits `json:"gasLimits"`
}
type NodeSendMessageResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type CanNodeBurnResponse struct {
	APIResponse
	CanBurn                bool            `json:"canBurn"`
	InsufficientBalance    bool            `json:"insufficientBalance"`
	InsufficientCollateral bool            `json:"insufficientCollateral"`
	GasLimits              gaslimit.Limits `json:"gasLimits"`
}
type NodeBurnResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type NodeSyncProgressResponse struct {
	APIResponse
	EcStatus ClientManagerStatus `json:"ecStatus"`
	BcStatus ClientManagerStatus `json:"bcStatus"`
}

type CanNodeClaimRplResponse struct {
	APIResponse
	RplAmount *big.Int        `json:"rplAmount"`
	GasLimits gaslimit.Limits `json:"gasLimits"`
}
type NodeClaimRplResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type NodeRewardsResponse struct {
	APIResponse
	NodeRegistrationTime        time.Time     `json:"nodeRegistrationTime"`
	RewardsInterval             time.Duration `json:"rewardsInterval"`
	LastCheckpoint              time.Time     `json:"lastCheckpoint"`
	Trusted                     bool          `json:"trusted"`
	Registered                  bool          `json:"registered"`
	EffectiveRplStake           float64       `json:"effectiveRplStake"`
	TotalRplStake               float64       `json:"totalRplStake"`
	TrustedRplBond              float64       `json:"trustedRplBond"`
	EstimatedRewards            float64       `json:"estimatedRewards"`
	CumulativeRplRewards        float64       `json:"cumulativeRplRewards"`
	CumulativeEthRewards        float64       `json:"cumulativeEthRewards"`
	EstimatedTrustedRplRewards  float64       `json:"estimatedTrustedRplRewards"`
	CumulativeTrustedRplRewards float64       `json:"cumulativeTrustedRplRewards"`
	UnclaimedRplRewards         float64       `json:"unclaimedRplRewards"`
	UnclaimedEthRewards         float64       `json:"unclaimedEthRewards"`
	UnclaimedTrustedRplRewards  float64       `json:"unclaimedTrustedRplRewards"`
	BeaconRewards               float64       `json:"beaconRewards"`
	TxHash                      common.Hash   `json:"txHash"`
}

type DepositContractInfoResponse struct {
	APIResponse
	RPDepositContract     common.Address `json:"rpDepositContract"`
	RPNetwork             uint64         `json:"rpNetwork"`
	BeaconDepositContract common.Address `json:"beaconDepositContract"`
	BeaconNetwork         uint64         `json:"beaconNetwork"`
	SufficientSync        bool           `json:"sufficientSync"`
}

type NodeSignResponse struct {
	APIResponse
	SignedData string `json:"signedData"`
}

type NodeIsFeeDistributorInitializedResponse struct {
	APIResponse
	IsInitialized bool `json:"isInitialized"`
}
type NodeInitializeFeeDistributorGasResponse struct {
	APIResponse
	Distributor common.Address  `json:"distributor"`
	GasLimits   gaslimit.Limits `json:"gasLimits"`
}
type NodeInitializeFeeDistributorResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}
type NodeCanDistributeResponse struct {
	APIResponse
	Balance   *big.Int        `json:"balance"`
	NodeShare float64         `json:"nodeShare"`
	GasLimits gaslimit.Limits `json:"gasLimits"`
}
type NodeDistributeResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type NodeGetRewardsInfoResponse struct {
	APIResponse
	Registered               bool                   `json:"registered"`
	ClaimedIntervals         []uint64               `json:"claimedIntervals"`
	UnclaimedIntervals       []rewards.IntervalInfo `json:"unclaimedIntervals"`
	InvalidIntervals         []rewards.IntervalInfo `json:"invalidIntervals"`
	RplStake                 *big.Int               `json:"rplStake"`
	RplPrice                 *big.Int               `json:"rplPrice"`
	ActiveMinipools          int                    `json:"activeMinipools"`
	ActiveMegapoolValidators int                    `json:"activeMegapoolValidators"`
	EthBorrowed              *big.Int               `json:"ethBorrowed"`
	EthBorrowLimit           *big.Int               `json:"ethBorrowLimit"`
	PendingBorrowAmount      *big.Int               `json:"pendingBorrowAmount"`
	BorrowedCollateralRatio  float64                `json:"borrowedCollateralRatio"`
	BondedCollateralRatio    float64                `json:"bondedCollateralRatio"`
}

type CanNodeClaimRewardsResponse struct {
	APIResponse
	GasLimits gaslimit.Limits `json:"gasLimits"`
}
type NodeClaimRewardsResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type CanNodeClaimAndStakeRewardsResponse struct {
	APIResponse
	GasLimits gaslimit.Limits `json:"gasLimits"`
}
type NodeClaimAndStakeRewardsResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type GetSmoothingPoolRegistrationStatusResponse struct {
	APIResponse
	NodeRegistered          bool          `json:"nodeRegistered"`
	TimeLeftUntilChangeable time.Duration `json:"timeLeftUntilChangeable"`
}
type CanSetSmoothingPoolRegistrationStatusResponse struct {
	APIResponse
	GasLimits gaslimit.Limits `json:"gasLimits"`
}
type SetSmoothingPoolRegistrationStatusResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}
type ResolveEnsNameResponse struct {
	APIResponse
	Address common.Address `json:"address"`
	EnsName string         `json:"ensName"`
}
type SnapshotProposal struct {
	Id            string    `json:"id"`
	Title         string    `json:"title"`
	Start         int64     `json:"start"`
	End           int64     `json:"end"`
	State         string    `json:"state"`
	Snapshot      int64     `json:"snapshot"`
	Author        string    `json:"author"`
	Choices       []string  `json:"choices"`
	Scores        []float64 `json:"scores"`
	ScoresTotal   float64   `json:"scores_total"`
	ScoresUpdated int64     `json:"scores_updated"`
	Quorum        float64   `json:"quorum"`
	Link          string    `json:"link"`
}
type SnapshotResponse struct {
	APIResponse
	Data struct {
		Proposals []SnapshotProposal `json:"proposals"`
	}
}
type SnapshotVotingPower struct {
	Data struct {
		Vp struct {
			Vp float64 `json:"vp"`
		} `json:"vp"`
	} `json:"data"`
}
type SnapshotProposalVote struct {
	Choice   interface{}    `json:"choice"`
	Voter    common.Address `json:"voter"`
	Proposal struct {
		Id    string `json:"id"`
		State string `json:"state"`
	} `json:"proposal"`
}
type SnapshotVotedProposals struct {
	APIResponse
	Data struct {
		Votes []SnapshotProposalVote `json:"votes"`
	} `json:"data"`
}
type SmoothingRewardsResponse struct {
	APIResponse
	EthBalance *big.Int `json:"eth_balance"`
}

type CheckCollateralResponse struct {
	APIResponse
	EthBorrowed            *big.Int `json:"ethBorrowed"`
	EthBorrowedLimit       *big.Int `json:"ethBorrowedLimit"`
	PendingBorrowAmount    *big.Int `json:"pendingBorrowAmount"`
	InsufficientCollateral bool     `json:"insufficientCollateral"`
}

type NodeEthBalanceResponse struct {
	APIResponse
	Balance *big.Int `json:"balance"`
}

type NodeAlertsResponse struct {
	APIResponse
	Alerts []NodeAlert `json:"alerts"`
}

type GetExpressTicketCountResponse struct {
	APIResponse
	Count uint64 `json:"count"`
}

type GetBondRequirementResponse struct {
	APIResponse
	BondRequirement *big.Int `json:"bondRequirement"`
}

type GetExpressTicketsProvisionedResponse struct {
	APIResponse
	Provisioned bool `json:"provisioned"`
}

type CanClaimRefundResponse struct {
	APIResponse
	CanClaim  bool            `json:"canClaim"`
	GasLimits gaslimit.Limits `json:"gasLimits"`
}
type ClaimRefundResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type CanReduceBondResponse struct {
	APIResponse
	CanReduceBond bool            `json:"canReduceBond"`
	NotEnoughBond bool            `json:"notEnoughBond"`
	GasLimits     gaslimit.Limits `json:"gasLimits"`
}
type ReduceBondResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type CanRepayDebtResponse struct {
	APIResponse
	CanRepay         bool            `json:"canRepay"`
	NotEnoughDebt    bool            `json:"notEnoughDebt"`
	NotEnoughBalance bool            `json:"notEnoughBalance"`
	GasLimits        gaslimit.Limits `json:"gasLimits"`
}
type RepayDebtResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type CanDissolveValidatorResponse struct {
	APIResponse
	CanDissolve   bool            `json:"canDissolve"`
	NotInPrestake bool            `json:"notInPrestake"`
	GasLimits     gaslimit.Limits `json:"gasLimits"`
}
type DissolveValidatorResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type CanDissolveWithProofResponse struct {
	APIResponse
	CanDissolve      bool            `json:"canDissolve"`
	NotInPrestake    bool            `json:"notInPrestake"`
	ValidCredentials bool            `json:"validCredentials"`
	GasLimits        gaslimit.Limits `json:"gasLimits"`
}
type DissolveWithProofResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type CanExitValidatorResponse struct {
	APIResponse
	CanExit       bool            `json:"canExit"`
	InvalidStatus bool            `json:"invalidStatus"`
	GasLimits     gaslimit.Limits `json:"gasLimits"`
}
type ExitValidatorResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type CanNotifyValidatorExitResponse struct {
	APIResponse
	CanExit          bool            `json:"canExit"`
	InvalidStatus    bool            `json:"invalidStatus"`
	AlreadyExiting   bool            `json:"alreadyExiting"`
	AlreadyExited    bool            `json:"alreadyExited"`
	ExitNotFinalized bool            `json:"exitNotFinalized"`
	GasLimits        gaslimit.Limits `json:"gasLimits"`
}
type NotifyValidatorExitResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type CanNotifyFinalBalanceResponse struct {
	APIResponse
	CanExit       bool            `json:"canExit"`
	InvalidStatus bool            `json:"invalidStatus"`
	GasLimits     gaslimit.Limits `json:"gasLimits"`
}
type NotifyFinalBalanceResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type CanStakeResponse struct {
	APIResponse
	CanStake      bool            `json:"canStake"`
	IndexNotFound bool            `json:"indexNotFound"`
	GasLimits     gaslimit.Limits `json:"gasLimits"`
}
type StakeResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type CanExitQueueResponse struct {
	APIResponse
	CanExit   bool            `json:"canExit"`
	GasLimits gaslimit.Limits `json:"gasLimits"`
}

type ExitQueueResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type CanClaimUnclaimedRewardsResponse struct {
	APIResponse
	CanClaim  bool            `json:"canClaim"`
	GasLimits gaslimit.Limits `json:"gasLimits"`
}

type ClaimUnclaimedRewardsResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}
