package api

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/tokens"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/shared/services/rewards"
	"github.com/rocket-pool/smartnode/shared/utils/rp"
)

type NodeStatusResponse struct {
	Status                            string          `json:"status"`
	Error                             string          `json:"error"`
	AccountAddress                    common.Address  `json:"accountAddress"`
	AccountAddressFormatted           string          `json:"accountAddressFormatted"`
	WithdrawalAddress                 common.Address  `json:"withdrawalAddress"`
	WithdrawalAddressFormatted        string          `json:"withdrawalAddressFormatted"`
	PendingWithdrawalAddress          common.Address  `json:"pendingWithdrawalAddress"`
	PendingWithdrawalAddressFormatted string          `json:"pendingWithdrawalAddressFormatted"`
	Registered                        bool            `json:"registered"`
	Trusted                           bool            `json:"trusted"`
	TimezoneLocation                  string          `json:"timezoneLocation"`
	AccountBalances                   tokens.Balances `json:"accountBalances"`
	WithdrawalBalances                tokens.Balances `json:"withdrawalBalances"`
	RplStake                          *big.Int        `json:"rplStake"`
	EffectiveRplStake                 *big.Int        `json:"effectiveRplStake"`
	MinimumRplStake                   *big.Int        `json:"minimumRplStake"`
	MaximumRplStake                   *big.Int        `json:"maximumRplStake"`
	MaximumStakeFraction              float64         `json:"maximumStakeFraction"`
	BorrowedCollateralRatio           float64         `json:"borrowedCollateralRatio"`
	BondedCollateralRatio             float64         `json:"bondedCollateralRatio"`
	PendingEffectiveRplStake          *big.Int        `json:"pendingEffectiveRplStake"`
	PendingMinimumRplStake            *big.Int        `json:"pendingMinimumRplStake"`
	PendingMaximumRplStake            *big.Int        `json:"pendingMaximumRplStake"`
	PendingBorrowedCollateralRatio    float64         `json:"pendingBorrowedCollateralRatio"`
	PendingBondedCollateralRatio      float64         `json:"pendingBondedCollateralRatio"`
	VotingDelegate                    common.Address  `json:"votingDelegate"`
	VotingDelegateFormatted           string          `json:"votingDelegateFormatted"`
	MinipoolLimit                     uint64          `json:"minipoolLimit"`
	EthMatched                        *big.Int        `json:"ethMatched"`
	EthMatchedLimit                   *big.Int        `json:"ethMatchedLimit"`
	PendingMatchAmount                *big.Int        `json:"pendingMatchAmount"`
	CreditBalance                     *big.Int        `json:"creditBalance"`
	MinipoolCounts                    struct {
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
	FeeRecipientInfo            rp.FeeRecipientInfo       `json:"feeRecipientInfo"`
	FeeDistributorBalance       *big.Int                  `json:"feeDistributorBalance"`
	PenalizedMinipools          map[common.Address]uint64 `json:"penalizedMinipools"`
	SnapshotResponse            struct {
		Error                   string                 `json:"error"`
		ProposalVotes           []SnapshotProposalVote `json:"proposalVotes"`
		ActiveSnapshotProposals []SnapshotProposal     `json:"activeSnapshotProposals"`
	} `json:"snapshotResponse"`
}

type CanRegisterNodeResponse struct {
	Status               string             `json:"status"`
	Error                string             `json:"error"`
	CanRegister          bool               `json:"canRegister"`
	AlreadyRegistered    bool               `json:"alreadyRegistered"`
	RegistrationDisabled bool               `json:"registrationDisabled"`
	GasInfo              rocketpool.GasInfo `json:"gasInfo"`
}
type RegisterNodeResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type CanSetNodeWithdrawalAddressResponse struct {
	Status  string             `json:"status"`
	Error   string             `json:"error"`
	CanSet  bool               ` json:"canSet"`
	GasInfo rocketpool.GasInfo `json:"gasInfo"`
}
type SetNodeWithdrawalAddressResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type CanConfirmNodeWithdrawalAddressResponse struct {
	Status     string             `json:"status"`
	Error      string             `json:"error"`
	CanConfirm bool               `json:"canConfirm"`
	GasInfo    rocketpool.GasInfo `json:"gasInfo"`
}
type ConfirmNodeWithdrawalAddressResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type GetNodeWithdrawalAddressResponse struct {
	Status  string         `json:"status"`
	Error   string         `json:"error"`
	Address common.Address `json:"address"`
}

type GetNodePendingWithdrawalAddressResponse struct {
	Status  string         `json:"status"`
	Error   string         `json:"error"`
	Address common.Address `json:"address"`
}

type CanSetNodeTimezoneResponse struct {
	Status  string             `json:"status"`
	Error   string             `json:"error"`
	CanSet  bool               `json:"canSet"`
	GasInfo rocketpool.GasInfo `json:"gasInfo"`
}
type SetNodeTimezoneResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type CanNodeSwapRplResponse struct {
	Status              string             `json:"status"`
	Error               string             `json:"error"`
	CanSwap             bool               `json:"canSwap"`
	InsufficientBalance bool               `json:"insufficientBalance"`
	GasInfo             rocketpool.GasInfo `json:"GasInfo"`
}
type NodeSwapRplApproveGasResponse struct {
	Status  string             `json:"status"`
	Error   string             `json:"error"`
	GasInfo rocketpool.GasInfo `json:"gasInfo"`
}
type NodeSwapRplApproveResponse struct {
	Status        string      `json:"status"`
	Error         string      `json:"error"`
	ApproveTxHash common.Hash `json:"approveTxHash"`
}
type NodeSwapRplSwapResponse struct {
	Status     string      `json:"status"`
	Error      string      `json:"error"`
	SwapTxHash common.Hash `json:"swapTxHash"`
}
type NodeSwapRplAllowanceResponse struct {
	Status    string   `json:"status"`
	Error     string   `json:"error"`
	Allowance *big.Int `json:"allowance"`
}

type CanNodeStakeRplResponse struct {
	Status              string             `json:"status"`
	Error               string             `json:"error"`
	CanStake            bool               `json:"canStake"`
	InsufficientBalance bool               `json:"insufficientBalance"`
	InConsensus         bool               `json:"inConsensus"`
	GasInfo             rocketpool.GasInfo `json:"gasInfo"`
}
type NodeStakeRplApproveGasResponse struct {
	Status  string             `json:"status"`
	Error   string             `json:"error"`
	GasInfo rocketpool.GasInfo `json:"gasInfo"`
}
type NodeStakeRplApproveResponse struct {
	Status        string      `json:"status"`
	Error         string      `json:"error"`
	ApproveTxHash common.Hash `json:"approveTxHash"`
}
type NodeStakeRplStakeResponse struct {
	Status      string      `json:"status"`
	Error       string      `json:"error"`
	StakeTxHash common.Hash `json:"stakeTxHash"`
}
type NodeStakeRplAllowanceResponse struct {
	Status    string   `json:"status"`
	Error     string   `json:"error"`
	Allowance *big.Int `json:"allowance"`
}

type CanSetStakeRplForAllowedResponse struct {
	Status  string             `json:"status"`
	Error   string             `json:"error"`
	CanSet  bool               `json:"canSet"`
	GasInfo rocketpool.GasInfo `json:"gasInfo"`
}
type SetStakeRplForAllowedResponse struct {
	Status    string      `json:"status"`
	Error     string      `json:"error"`
	SetTxHash common.Hash `json:"setTxHash"`
}

type CanNodeWithdrawRplResponse struct {
	Status                       string             `json:"status"`
	Error                        string             `json:"error"`
	CanWithdraw                  bool               `json:"canWithdraw"`
	InsufficientBalance          bool               `json:"insufficientBalance"`
	BelowMaxRPLStake             bool               `json:"belowMaxRPLStake"`
	MinipoolsUndercollateralized bool               `json:"minipoolsUndercollateralized"`
	WithdrawalDelayActive        bool               `json:"withdrawalDelayActive"`
	InConsensus                  bool               `json:"inConsensus"`
	GasInfo                      rocketpool.GasInfo `json:"gasInfo"`
}
type NodeWithdrawRplResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type CanNodeDepositResponse struct {
	Status                           string             `json:"status"`
	Error                            string             `json:"error"`
	CanDeposit                       bool               `json:"canDeposit"`
	CreditBalance                    *big.Int           `json:"creditBalance"`
	DepositBalance                   *big.Int           `json:"depositBalance"`
	CanUseCredit                     bool               `json:"canUseCredit"`
	NodeBalance                      *big.Int           `json:"nodeBalance"`
	InsufficientBalance              bool               `json:"insufficientBalance"`
	InsufficientBalanceWithoutCredit bool               `json:"insufficientBalanceWithoutCredit"`
	InsufficientRplStake             bool               `json:"insufficientRplStake"`
	InvalidAmount                    bool               `json:"invalidAmount"`
	UnbondedMinipoolsAtMax           bool               `json:"unbondedMinipoolsAtMax"`
	DepositDisabled                  bool               `json:"depositDisabled"`
	InConsensus                      bool               `json:"inConsensus"`
	MinipoolAddress                  common.Address     `json:"minipoolAddress"`
	GasInfo                          rocketpool.GasInfo `json:"gasInfo"`
}
type NodeDepositResponse struct {
	Status          string                  `json:"status"`
	Error           string                  `json:"error"`
	TxHash          common.Hash             `json:"txHash"`
	MinipoolAddress common.Address          `json:"minipoolAddress"`
	ValidatorPubkey rptypes.ValidatorPubkey `json:"validatorPubkey"`
	ScrubPeriod     time.Duration           `json:"scrubPeriod"`
}

type CanCreateVacantMinipoolResponse struct {
	Status               string             `json:"status"`
	Error                string             `json:"error"`
	CanDeposit           bool               `json:"canDeposit"`
	InsufficientRplStake bool               `json:"insufficientRplStake"`
	InvalidAmount        bool               `json:"invalidAmount"`
	DepositDisabled      bool               `json:"depositDisabled"`
	MinipoolAddress      common.Address     `json:"minipoolAddress"`
	GasInfo              rocketpool.GasInfo `json:"gasInfo"`
}
type CreateVacantMinipoolResponse struct {
	Status                string         `json:"status"`
	Error                 string         `json:"error"`
	TxHash                common.Hash    `json:"txHash"`
	MinipoolAddress       common.Address `json:"minipoolAddress"`
	ScrubPeriod           time.Duration  `json:"scrubPeriod"`
	WithdrawalCredentials common.Hash    `json:"withdrawalCredentials"`
}

type CanNodeSendResponse struct {
	Status              string             `json:"status"`
	Error               string             `json:"error"`
	Balance             *big.Int           `json:"balance"`
	TokenName           string             `json:"name"`
	TokenSymbol         string             `json:"symbol"`
	CanSend             bool               `json:"canSend"`
	InsufficientBalance bool               `json:"insufficientBalance"`
	GasInfo             rocketpool.GasInfo `json:"gasInfo"`
}
type NodeSendResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type CanNodeSendMessageResponse struct {
	Status  string             `json:"status"`
	Error   string             `json:"error"`
	GasInfo rocketpool.GasInfo `json:"gasInfo"`
}
type NodeSendMessageResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type CanNodeBurnResponse struct {
	Status                 string             `json:"status"`
	Error                  string             `json:"error"`
	CanBurn                bool               `json:"canBurn"`
	InsufficientBalance    bool               `json:"insufficientBalance"`
	InsufficientCollateral bool               `json:"insufficientCollateral"`
	GasInfo                rocketpool.GasInfo `json:"gasInfo"`
}
type NodeBurnResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type NodeSyncProgressResponse struct {
	Status   string              `json:"status"`
	Error    string              `json:"error"`
	EcStatus ClientManagerStatus `json:"ecStatus"`
	BcStatus ClientManagerStatus `json:"bcStatus"`
}

type CanNodeClaimRplResponse struct {
	Status    string             `json:"status"`
	Error     string             `json:"error"`
	RplAmount *big.Int           `json:"rplAmount"`
	GasInfo   rocketpool.GasInfo `json:"gasInfo"`
}
type NodeClaimRplResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type NodeRewardsResponse struct {
	Status                      string        `json:"status"`
	Error                       string        `json:"error"`
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
	Status                string         `json:"status"`
	Error                 string         `json:"error"`
	RPDepositContract     common.Address `json:"rpDepositContract"`
	RPNetwork             uint64         `json:"rpNetwork"`
	BeaconDepositContract common.Address `json:"beaconDepositContract"`
	BeaconNetwork         uint64         `json:"beaconNetwork"`
	SufficientSync        bool           `json:"sufficientSync"`
}

type NodeSignResponse struct {
	Status     string `json:"status"`
	Error      string `json:"error"`
	SignedData string `json:"signedData"`
}

type EstimateSetSnapshotDelegateGasResponse struct {
	Status  string             `json:"status"`
	Error   string             `json:"error"`
	GasInfo rocketpool.GasInfo `json:"gasInfo"`
}

type SetSnapshotDelegateResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type EstimateClearSnapshotDelegateGasResponse struct {
	Status  string             `json:"status"`
	Error   string             `json:"error"`
	GasInfo rocketpool.GasInfo `json:"gasInfo"`
}

type ClearSnapshotDelegateResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type NodeIsFeeDistributorInitializedResponse struct {
	Status        string `json:"status"`
	Error         string `json:"error"`
	IsInitialized bool   `json:"isInitialized"`
}
type NodeInitializeFeeDistributorGasResponse struct {
	Status      string             `json:"status"`
	Error       string             `json:"error"`
	Distributor common.Address     `json:"distributor"`
	GasInfo     rocketpool.GasInfo `json:"gasInfo"`
}
type NodeInitializeFeeDistributorResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}
type NodeCanDistributeResponse struct {
	Status    string             `json:"status"`
	Error     string             `json:"error"`
	Balance   *big.Int           `json:"balance"`
	NodeShare float64            `json:"nodeShare"`
	GasInfo   rocketpool.GasInfo `json:"gasInfo"`
}
type NodeDistributeResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type NodeGetRewardsInfoResponse struct {
	Status                  string                 `json:"status"`
	Error                   string                 `json:"error"`
	Registered              bool                   `json:"registered"`
	ClaimedIntervals        []uint64               `json:"claimedIntervals"`
	UnclaimedIntervals      []rewards.IntervalInfo `json:"unclaimedIntervals"`
	InvalidIntervals        []rewards.IntervalInfo `json:"invalidIntervals"`
	RplStake                *big.Int               `json:"rplStake"`
	RplPrice                *big.Int               `json:"rplPrice"`
	ActiveMinipools         int                    `json:"activeMinipools"`
	EffectiveRplStake       *big.Int               `json:"effectiveRplStake"`
	MinimumRplStake         *big.Int               `json:"minimumRplStake"`
	EthMatched              *big.Int               `json:"ethMatched"`
	EthMatchedLimit         *big.Int               `json:"ethMatchedLimit"`
	PendingMatchAmount      *big.Int               `json:"pendingMatchAmount"`
	BorrowedCollateralRatio float64                `json:"borrowedCollateralRatio"`
	BondedCollateralRatio   float64                `json:"bondedCollateralRatio"`
}

type CanNodeClaimRewardsResponse struct {
	Status  string             `json:"status"`
	Error   string             `json:"error"`
	GasInfo rocketpool.GasInfo `json:"gasInfo"`
}
type NodeClaimRewardsResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type CanNodeClaimAndStakeRewardsResponse struct {
	Status  string             `json:"status"`
	Error   string             `json:"error"`
	GasInfo rocketpool.GasInfo `json:"gasInfo"`
}
type NodeClaimAndStakeRewardsResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type GetSmoothingPoolRegistrationStatusResponse struct {
	Status                  string        `json:"status"`
	Error                   string        `json:"error"`
	NodeRegistered          bool          `json:"nodeRegistered"`
	TimeLeftUntilChangeable time.Duration `json:"timeLeftUntilChangeable"`
}
type CanSetSmoothingPoolRegistrationStatusResponse struct {
	Status  string             `json:"status"`
	Error   string             `json:"error"`
	GasInfo rocketpool.GasInfo `json:"gasInfo"`
}
type SetSmoothingPoolRegistrationStatusResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}
type ResolveEnsNameResponse struct {
	Status  string         `json:"status"`
	Error   string         `json:"error"`
	Address common.Address `json:"address"`
	EnsName string         `json:"ensName"`
}
type SnapshotProposal struct {
	Id            string    `json:"id"`
	Title         string    `json:"title"`
	Start         int64     `json:"start"`
	End           int64     `json:"end"`
	State         string    `json:"state"`
	Snapshot      string    `json:"snapshot"`
	Author        string    `json:"author"`
	Choices       []string  `json:"choices"`
	Scores        []float64 `json:"scores"`
	ScoresTotal   float64   `json:"scores_total"`
	ScoresUpdated int64     `json:"scores_updated"`
	Quorum        float64   `json:"quorum"`
	Link          string    `json:"link"`
}
type SnapshotResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
	Data   struct {
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
	Status string `json:"status"`
	Error  string `json:"error"`
	Data   struct {
		Votes []SnapshotProposalVote `json:"votes"`
	} `json:"data"`
}
type SmoothingRewardsResponse struct {
	Status     string   `json:"status"`
	Error      string   `json:"error"`
	EthBalance *big.Int `json:"eth_balance"`
}

type CheckCollateralResponse struct {
	Status                 string   `json:"status"`
	Error                  string   `json:"error"`
	EthMatched             *big.Int `json:"ethMatched"`
	EthMatchedLimit        *big.Int `json:"ethMatchedLimit"`
	PendingMatchAmount     *big.Int `json:"pendingMatchAmount"`
	InsufficientCollateral bool     `json:"insufficientCollateral"`
}

type NodeEthBalanceResponse struct {
	Status  string   `json:"status"`
	Error   string   `json:"error"`
	Balance *big.Int `json:"balance"`
}
