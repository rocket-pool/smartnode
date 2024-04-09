package api

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/eth"

	sharedtypes "github.com/rocket-pool/smartnode/v2/shared/types"
)

// Info for the node's fee recipient
type FeeRecipientInfo struct {
	SmoothingPoolAddress  common.Address `json:"smoothingPoolAddress"`
	FeeDistributorAddress common.Address `json:"feeDistributorAddress"`
	IsInSmoothingPool     bool           `json:"isInSmoothingPool"`
	IsInOptOutCooldown    bool           `json:"isInOptOutCooldown"`
	OptOutEpoch           uint64         `json:"optOutEpoch"`
}

type NodeStatusData struct {
	Warning                                  string         `json:"warning"`
	AccountAddress                           common.Address `json:"accountAddress"`
	AccountAddressFormatted                  string         `json:"accountAddressFormatted"`
	PrimaryWithdrawalAddress                 common.Address `json:"primaryWithdrawalAddress"`
	PrimaryWithdrawalAddressFormatted        string         `json:"primaryWithdrawalAddressFormatted"`
	PendingPrimaryWithdrawalAddress          common.Address `json:"pendingWithdrawalAddress"`
	PendingPrimaryWithdrawalAddressFormatted string         `json:"pendingWithdrawalAddressFormatted"`
	IsRplWithdrawalAddressSet                bool           `json:"isRplWithdrawalAddressSet"`
	RplWithdrawalAddress                     common.Address `json:"rplWithdrawalAddress"`
	RplWithdrawalAddressFormatted            string         `json:"rplWithdrawalAddressFormatted"`
	PendingRplWithdrawalAddress              common.Address `json:"pendingRplWithdrawalAddress"`
	PendingRplWithdrawalAddressFormatted     string         `json:"pendingRplWithdrawalAddressFormatted"`
	IsRplLockingAllowed                      bool           `json:"isRplLockingAllowed"`
	RplLocked                                *big.Int       `json:"rplLocked"`
	Registered                               bool           `json:"registered"`
	Trusted                                  bool           `json:"trusted"`
	TimezoneLocation                         string         `json:"timezoneLocation"`
	NodeBalances                             struct {
		Eth   *big.Int `json:"eth"`
		Reth  *big.Int `json:"reth"`
		Rpl   *big.Int `json:"rpl"`
		Fsrpl *big.Int `json:"fsrpl"`
	} `json:"nodeBalances"`
	PrimaryWithdrawalBalances struct {
		Eth   *big.Int `json:"eth"`
		Reth  *big.Int `json:"reth"`
		Rpl   *big.Int `json:"rpl"`
		Fsrpl *big.Int `json:"fsrpl"`
	} `json:"primaryWithdrawalBalances"`
	RplWithdrawalBalances struct {
		Eth   *big.Int `json:"eth"`
		Reth  *big.Int `json:"reth"`
		Rpl   *big.Int `json:"rpl"`
		Fsrpl *big.Int `json:"fsrpl"`
	} `json:"rplWithdrawalBalances"`
	RplStake                          *big.Int       `json:"rplStake"`
	EffectiveRplStake                 *big.Int       `json:"effectiveRplStake"`
	MinimumRplStake                   *big.Int       `json:"minimumRplStake"`
	MaximumRplStake                   *big.Int       `json:"maximumRplStake"`
	MaximumStakeFraction              *big.Int       `json:"maximumStakeFraction"`
	BorrowedCollateralRatio           float64        `json:"borrowedCollateralRatio"`
	BondedCollateralRatio             float64        `json:"bondedCollateralRatio"`
	PendingEffectiveRplStake          *big.Int       `json:"pendingEffectiveRplStake"`
	PendingMinimumRplStake            *big.Int       `json:"pendingMinimumRplStake"`
	PendingMaximumRplStake            *big.Int       `json:"pendingMaximumRplStake"`
	PendingBorrowedCollateralRatio    float64        `json:"pendingBorrowedCollateralRatio"`
	PendingBondedCollateralRatio      float64        `json:"pendingBondedCollateralRatio"`
	SnapshotVotingDelegate            common.Address `json:"votingDelegate"`
	SnapshotVotingDelegateFormatted   string         `json:"votingDelegateFormatted"`
	IsVotingInitialized               bool           `json:"isVotingInitialized"`
	OnchainVotingDelegate             common.Address `json:"onchainVotingDelegate"`
	OnchainVotingDelegateFormatted    string         `json:"onchainVotingDelegateFormatted"`
	MinipoolLimit                     uint64         `json:"minipoolLimit"`
	EthMatched                        *big.Int       `json:"ethMatched"`
	EthMatchedLimit                   *big.Int       `json:"ethMatchedLimit"`
	PendingMatchAmount                *big.Int       `json:"pendingMatchAmount"`
	CreditBalance                     *big.Int       `json:"creditBalance"`
	CreditAndEthOnBehalfBalance       *big.Int       `json:"creditAndEthOnBehalfBalance"`
	EthOnBehalfBalance                *big.Int       `json:"ethOnBehalfBalance"`
	UsableCreditAndEthOnBehalfBalance *big.Int       `json:"usableCreditAndEthOnBehalfBalance"`
	MinipoolCounts                    struct {
		Total           int `json:"total"`
		Initialized     int `json:"initialized"`
		Prelaunch       int `json:"prelaunch"`
		Staking         int `json:"staking"`
		Withdrawable    int `json:"withdrawable"`
		Dissolved       int `json:"dissolved"`
		RefundAvailable int `json:"refundAvailable"`
		Finalised       int `json:"finalised"`
	} `json:"minipoolCounts"`
	IsFeeDistributorInitialized bool                      `json:"isFeeDistributorInitialized"`
	FeeRecipientInfo            FeeRecipientInfo          `json:"feeRecipientInfo"`
	FeeDistributorBalance       *big.Int                  `json:"feeDistributorBalance"`
	PenalizedMinipools          map[common.Address]uint64 `json:"penalizedMinipools"`
	SnapshotResponse            struct {
		Error                   string                          `json:"error"`
		ActiveSnapshotProposals []*sharedtypes.SnapshotProposal `json:"activeSnapshotProposals"`
	} `json:"snapshotResponse"`
	Alerts []NodeAlert `json:"alerts"`
}

type NodeRegisterData struct {
	CanRegister          bool                 `json:"canRegister"`
	AlreadyRegistered    bool                 `json:"alreadyRegistered"`
	RegistrationDisabled bool                 `json:"registrationDisabled"`
	TxInfo               *eth.TransactionInfo `json:"txInfo"`
}

type NodeSetRplLockingAllowedData struct {
	CanSet              bool                 `json:"canSet"`
	DifferentRplAddress bool                 `json:"differentRplAddress"`
	TxInfo              *eth.TransactionInfo `json:"txInfo"`
}

type NodeSetPrimaryWithdrawalAddressData struct {
	CanSet            bool                 `json:"canSet"`
	AddressAlreadySet bool                 `json:"addressAlreadySet"`
	TxInfo            *eth.TransactionInfo `json:"txInfo"`
}

type NodeConfirmPrimaryWithdrawalAddressData struct {
	CanConfirm              bool                 `json:"canConfirm"`
	IncorrectPendingAddress bool                 `json:"incorrectPendingAddress"`
	TxInfo                  *eth.TransactionInfo `json:"txInfo"`
}

type NodeSetRplWithdrawalAddressData struct {
	CanSet                bool                 `json:"canSet"`
	PrimaryAddressDiffers bool                 `json:"primaryAddressDiffers"`
	RplAddressDiffers     bool                 `json:"rplAddressDiffers"`
	RplStake              *big.Int             `json:"rplStake"`
	TxInfo                *eth.TransactionInfo `json:"txInfo"`
}

type NodeConfirmRplWithdrawalAddressData struct {
	CanConfirm              bool                 `json:"canConfirm"`
	IncorrectPendingAddress bool                 `json:"incorrectPendingAddress"`
	TxInfo                  *eth.TransactionInfo `json:"txInfo"`
}

type NodeSwapRplData struct {
	CanSwap             bool                 `json:"canSwap"`
	InsufficientBalance bool                 `json:"insufficientBalance"`
	Allowance           *big.Int             `json:"allowance"`
	ApproveTxInfo       *eth.TransactionInfo `json:"approveTxInfo"`
	SwapTxInfo          *eth.TransactionInfo `json:"swapTxInfo"`
}

type NodeStakeRplData struct {
	CanStake            bool                 `json:"canStake"`
	InsufficientBalance bool                 `json:"insufficientBalance"`
	Allowance           *big.Int             `json:"allowance"`
	ApproveTxInfo       *eth.TransactionInfo `json:"approveTxInfo"`
	StakeTxInfo         *eth.TransactionInfo `json:"stakeTxInfo"`
}

type NodeSetStakeRplForAllowedData struct {
	CanSet bool                 `json:"canSet"`
	TxInfo *eth.TransactionInfo `json:"txInfo"`
}

type NodeWithdrawRplData struct {
	CanWithdraw                      bool                 `json:"canWithdraw"`
	InsufficientBalance              bool                 `json:"insufficientBalance"`
	BelowMaxRplStake                 bool                 `json:"belowMaxRplStake"`
	MinipoolsUndercollateralized     bool                 `json:"minipoolsUndercollateralized"`
	WithdrawalDelayActive            bool                 `json:"withdrawalDelayActive"`
	HasDifferentRplWithdrawalAddress bool                 `json:"hasDifferentRPLWithdrawalAddress"`
	TxInfo                           *eth.TransactionInfo `json:"txInfo"`
}

type NodeWithdrawEthData struct {
	CanWithdraw                          bool                 `json:"canWithdraw"`
	InsufficientBalance                  bool                 `json:"insufficientBalance"`
	HasDifferentPrimaryWithdrawalAddress bool                 `json:"hasDifferentWithdrawalAddress"`
	TxInfo                               *eth.TransactionInfo `json:"txInfo"`
}

type NodeDepositData struct {
	CanDeposit                       bool                   `json:"canDeposit"`
	CreditBalance                    *big.Int               `json:"creditBalance"`
	DepositBalance                   *big.Int               `json:"depositBalance"`
	CanUseCredit                     bool                   `json:"canUseCredit"`
	NodeBalance                      *big.Int               `json:"nodeBalance"`
	InsufficientBalance              bool                   `json:"insufficientBalance"`
	InsufficientBalanceWithoutCredit bool                   `json:"insufficientBalanceWithoutCredit"`
	InsufficientRplStake             bool                   `json:"insufficientRplStake"`
	InvalidAmount                    bool                   `json:"invalidAmount"`
	UnbondedMinipoolsAtMax           bool                   `json:"unbondedMinipoolsAtMax"`
	DepositDisabled                  bool                   `json:"depositDisabled"`
	InConsensus                      bool                   `json:"inConsensus"`
	MinipoolAddress                  common.Address         `json:"minipoolAddress"`
	ValidatorPubkey                  beacon.ValidatorPubkey `json:"validatorPubkey"`
	Index                            uint64                 `json:"index"`
	ScrubPeriod                      time.Duration          `json:"scrubPeriod"`
	TxInfo                           *eth.TransactionInfo   `json:"txInfo"`
}

type NodeCreateVacantMinipoolData struct {
	CanDeposit            bool                 `json:"canDeposit"`
	InsufficientRplStake  bool                 `json:"insufficientRplStake"`
	InvalidAmount         bool                 `json:"invalidAmount"`
	DepositDisabled       bool                 `json:"depositDisabled"`
	MinipoolAddress       common.Address       `json:"minipoolAddress"`
	ScrubPeriod           time.Duration        `json:"scrubPeriod"`
	WithdrawalCredentials common.Hash          `json:"withdrawalCredentials"`
	TxInfo                *eth.TransactionInfo `json:"txInfo"`
}

type NodeSendData struct {
	Balance             *big.Int             `json:"balance"`
	TokenName           string               `json:"name"`
	TokenSymbol         string               `json:"symbol"`
	CanSend             bool                 `json:"canSend"`
	InsufficientBalance bool                 `json:"insufficientBalance"`
	TxInfo              *eth.TransactionInfo `json:"txInfo"`
}

type NodeSendMessageData struct {
	TxInfo *eth.TransactionInfo `json:"txInfo"`
}

type NodeBurnData struct {
	CanBurn                bool                 `json:"canBurn"`
	InsufficientBalance    bool                 `json:"insufficientBalance"`
	InsufficientCollateral bool                 `json:"insufficientCollateral"`
	TxInfo                 *eth.TransactionInfo `json:"txInfo"`
}

type NodeRewardsData struct {
	NodeRegistrationTime        time.Time            `json:"nodeRegistrationTime"`
	RewardsInterval             time.Duration        `json:"rewardsInterval"`
	LastCheckpoint              time.Time            `json:"lastCheckpoint"`
	Trusted                     bool                 `json:"trusted"`
	Registered                  bool                 `json:"registered"`
	EffectiveRplStake           float64              `json:"effectiveRplStake"`
	TotalRplStake               float64              `json:"totalRplStake"`
	TrustedRplBond              float64              `json:"trustedRplBond"`
	EstimatedRewards            float64              `json:"estimatedRewards"`
	CumulativeRplRewards        float64              `json:"cumulativeRplRewards"`
	CumulativeEthRewards        float64              `json:"cumulativeEthRewards"`
	EstimatedTrustedRplRewards  float64              `json:"estimatedTrustedRplRewards"`
	CumulativeTrustedRplRewards float64              `json:"cumulativeTrustedRplRewards"`
	UnclaimedRplRewards         float64              `json:"unclaimedRplRewards"`
	UnclaimedEthRewards         float64              `json:"unclaimedEthRewards"`
	UnclaimedTrustedRplRewards  float64              `json:"unclaimedTrustedRplRewards"`
	BeaconRewards               float64              `json:"beaconRewards"`
	TxInfo                      *eth.TransactionInfo `json:"txInfo"`
}

type NodeSignData struct {
	SignedData string `json:"signedData"`
}

type NodeInitializeFeeDistributorData struct {
	CanInitialize bool                 `json:"canInitialize"`
	IsInitialized bool                 `json:"isInitialized"`
	Distributor   common.Address       `json:"distributor"`
	TxInfo        *eth.TransactionInfo `json:"txInfo"`
}

type NodeDistributeData struct {
	CanDistribute bool                 `json:"canDistribute"`
	NoBalance     bool                 `json:"noBalance"`
	IsInitialized bool                 `json:"isInitialized"`
	Balance       *big.Int             `json:"balance"`
	NodeShare     *big.Int             `json:"nodeShare"`
	TxInfo        *eth.TransactionInfo `json:"txInfo"`
}

type NodeGetRewardsInfoData struct {
	ClaimedIntervals        []uint64                   `json:"claimedIntervals"`
	UnclaimedIntervals      []sharedtypes.IntervalInfo `json:"unclaimedIntervals"`
	InvalidIntervals        []sharedtypes.IntervalInfo `json:"invalidIntervals"`
	RplStake                *big.Int                   `json:"rplStake"`
	RplPrice                *big.Int                   `json:"rplPrice"`
	ActiveMinipools         uint64                     `json:"activeMinipools"`
	EffectiveRplStake       *big.Int                   `json:"effectiveRplStake"`
	MinimumRplStake         *big.Int                   `json:"minimumRplStake"`
	EthMatched              *big.Int                   `json:"ethMatched"`
	EthMatchedLimit         *big.Int                   `json:"ethMatchedLimit"`
	PendingMatchAmount      *big.Int                   `json:"pendingMatchAmount"`
	BorrowedCollateralRatio float64                    `json:"borrowedCollateralRatio"`
	BondedCollateralRatio   float64                    `json:"bondedCollateralRatio"`
}

type NodeSetSmoothingPoolRegistrationStatusData struct {
	NodeRegistered          bool                 `json:"nodeRegistered"`
	CanChange               bool                 `json:"canChange"`
	TimeLeftUntilChangeable time.Duration        `json:"timeLeftUntilChangeable"`
	TxInfo                  *eth.TransactionInfo `json:"txInfo"`
}

type NodeResolveEnsData struct {
	Address       common.Address `json:"address"`
	EnsName       string         `json:"ensName"`
	FormattedName string         `json:"formattedName"`
}

type NodeGetSnapshotVotingPowerData struct {
	VotingPower float64 `json:"votingPower"`
}

type NodeGetSnapshotProposalsData struct {
	Proposals []*sharedtypes.SnapshotProposal `json:"proposals"`
}

type NodeSmoothingRewardsData struct {
	EthBalance *big.Int `json:"eth_balance"`
}

type NodeCheckCollateralData struct {
	EthMatched             *big.Int `json:"ethMatched"`
	EthMatchedLimit        *big.Int `json:"ethMatchedLimit"`
	PendingMatchAmount     *big.Int `json:"pendingMatchAmount"`
	InsufficientCollateral bool     `json:"insufficientCollateral"`
}

type NodeBalanceData struct {
	Balance *big.Int `json:"balance"`
}

type NodeAlertsData struct {
	// TODO: change to GettableAlerts
	Message string `json:"message"`
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
	value, ok := n.Annotations["alertname"]
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
	const (
		colorReset  string = "\033[0m"
		colorRed    string = "\033[31m"
		colorYellow string = "\033[33m"
	)
	suppressed := ""
	if n.IsSuppressed() {
		suppressed = " (suppressed)"
	}
	alertColor := colorYellow
	if n.Severity() == "critical" {
		alertColor = colorRed
	}
	return fmt.Sprintf("%s%s%s%s %s: %s", alertColor, n.Severity(), suppressed, colorReset, n.Summary(), n.Description())
}
