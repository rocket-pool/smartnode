package api

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/core"
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

type NodeRegisterResponse struct {
	Status               string                `json:"status"`
	Error                string                `json:"error"`
	CanRegister          bool                  `json:"canRegister"`
	AlreadyRegistered    bool                  `json:"alreadyRegistered"`
	RegistrationDisabled bool                  `json:"registrationDisabled"`
	TxInfo               *core.TransactionInfo `json:"txInfo"`
}

type CanSetNodeWithdrawalAddressResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
	CanSet bool   ` json:"canSet"`
}
type SetNodeWithdrawalAddressResponse struct {
	Status string                `json:"status"`
	Error  string                `json:"error"`
	TxInfo *core.TransactionInfo `json:"txInfo"`
}

type CanConfirmNodeWithdrawalAddressResponse struct {
	Status     string `json:"status"`
	Error      string `json:"error"`
	CanConfirm bool   `json:"canConfirm"`
}
type ConfirmNodeWithdrawalAddressResponse struct {
	Status string                `json:"status"`
	Error  string                `json:"error"`
	TxInfo *core.TransactionInfo `json:"txInfo"`
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
	Status string `json:"status"`
	Error  string `json:"error"`
	CanSet bool   `json:"canSet"`
}
type SetNodeTimezoneResponse struct {
	Status string                `json:"status"`
	Error  string                `json:"error"`
	TxInfo *core.TransactionInfo `json:"txInfo"`
}

type CanNodeSwapRplResponse struct {
	Status              string `json:"status"`
	Error               string `json:"error"`
	CanSwap             bool   `json:"canSwap"`
	InsufficientBalance bool   `json:"insufficientBalance"`
}
type NodeSwapRplApproveGasResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}
type NodeSwapRplApproveResponse struct {
	Status string                `json:"status"`
	Error  string                `json:"error"`
	TxInfo *core.TransactionInfo `json:"txInfo"`
}
type NodeSwapRplSwapResponse struct {
	Status string                `json:"status"`
	Error  string                `json:"error"`
	TxInfo *core.TransactionInfo `json:"txInfo"`
}
type NodeSwapRplAllowanceResponse struct {
	Status    string   `json:"status"`
	Error     string   `json:"error"`
	Allowance *big.Int `json:"allowance"`
}

type CanNodeStakeRplResponse struct {
	Status              string `json:"status"`
	Error               string `json:"error"`
	CanStake            bool   `json:"canStake"`
	InsufficientBalance bool   `json:"insufficientBalance"`
	InConsensus         bool   `json:"inConsensus"`
}
type NodeStakeRplApproveGasResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}
type NodeStakeRplApproveResponse struct {
	Status string                `json:"status"`
	Error  string                `json:"error"`
	TxInfo *core.TransactionInfo `json:"txInfo"`
}
type NodeStakeRplStakeResponse struct {
	Status string                `json:"status"`
	Error  string                `json:"error"`
	TxInfo *core.TransactionInfo `json:"txInfo"`
}
type NodeStakeRplAllowanceResponse struct {
	Status    string   `json:"status"`
	Error     string   `json:"error"`
	Allowance *big.Int `json:"allowance"`
}

type CanSetStakeRplForAllowedResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
	CanSet bool   `json:"canSet"`
}
type SetStakeRplForAllowedResponse struct {
	Status string                `json:"status"`
	Error  string                `json:"error"`
	TxInfo *core.TransactionInfo `json:"txInfo"`
}

type CanNodeWithdrawRplResponse struct {
	Status                       string `json:"status"`
	Error                        string `json:"error"`
	CanWithdraw                  bool   `json:"canWithdraw"`
	InsufficientBalance          bool   `json:"insufficientBalance"`
	MinipoolsUndercollateralized bool   `json:"minipoolsUndercollateralized"`
	WithdrawalDelayActive        bool   `json:"withdrawalDelayActive"`
	InConsensus                  bool   `json:"inConsensus"`
}
type NodeWithdrawRplResponse struct {
	Status string                `json:"status"`
	Error  string                `json:"error"`
	TxInfo *core.TransactionInfo `json:"txInfo"`
}

type NodeDepositResponse struct {
	Status                           string                  `json:"status"`
	Error                            string                  `json:"error"`
	CanDeposit                       bool                    `json:"canDeposit"`
	CreditBalance                    *big.Int                `json:"creditBalance"`
	DepositBalance                   *big.Int                `json:"depositBalance"`
	CanUseCredit                     bool                    `json:"canUseCredit"`
	NodeBalance                      *big.Int                `json:"nodeBalance"`
	InsufficientBalance              bool                    `json:"insufficientBalance"`
	InsufficientBalanceWithoutCredit bool                    `json:"insufficientBalanceWithoutCredit"`
	InsufficientRplStake             bool                    `json:"insufficientRplStake"`
	InvalidAmount                    bool                    `json:"invalidAmount"`
	UnbondedMinipoolsAtMax           bool                    `json:"unbondedMinipoolsAtMax"`
	DepositDisabled                  bool                    `json:"depositDisabled"`
	InConsensus                      bool                    `json:"inConsensus"`
	MinipoolAddress                  common.Address          `json:"minipoolAddress"`
	ValidatorPubkey                  rptypes.ValidatorPubkey `json:"validatorPubkey"`
	ScrubPeriod                      time.Duration           `json:"scrubPeriod"`
	TxInfo                           *core.TransactionInfo   `json:"txInfo"`
}

type NodeCreateVacantMinipoolResponse struct {
	Status                string                `json:"status"`
	Error                 string                `json:"error"`
	CanDeposit            bool                  `json:"canDeposit"`
	InsufficientRplStake  bool                  `json:"insufficientRplStake"`
	InvalidAmount         bool                  `json:"invalidAmount"`
	DepositDisabled       bool                  `json:"depositDisabled"`
	MinipoolAddress       common.Address        `json:"minipoolAddress"`
	ScrubPeriod           time.Duration         `json:"scrubPeriod"`
	WithdrawalCredentials common.Hash           `json:"withdrawalCredentials"`
	TxInfo                *core.TransactionInfo `json:"txInfo"`
}

type CanNodeSendResponse struct {
	Status              string   `json:"status"`
	Error               string   `json:"error"`
	Balance             *big.Int `json:"balance"`
	TokenName           string   `json:"name"`
	TokenSymbol         string   `json:"symbol"`
	CanSend             bool     `json:"canSend"`
	InsufficientBalance bool     `json:"insufficientBalance"`
}
type NodeSendResponse struct {
	Status string                `json:"status"`
	Error  string                `json:"error"`
	TxInfo *core.TransactionInfo `json:"txInfo"`
}

type CanNodeSendMessageResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}
type NodeSendMessageResponse struct {
	Status string                `json:"status"`
	Error  string                `json:"error"`
	TxInfo *core.TransactionInfo `json:"txInfo"`
}

type NodeBurnResponse struct {
	Status                 string                `json:"status"`
	Error                  string                `json:"error"`
	CanBurn                bool                  `json:"canBurn"`
	InsufficientBalance    bool                  `json:"insufficientBalance"`
	InsufficientCollateral bool                  `json:"insufficientCollateral"`
	TxInfo                 *core.TransactionInfo `json:"txInfo"`
}

type NodeSyncProgressResponse struct {
	Status   string              `json:"status"`
	Error    string              `json:"error"`
	EcStatus ClientManagerStatus `json:"ecStatus"`
	BcStatus ClientManagerStatus `json:"bcStatus"`
}

type CanNodeClaimRplResponse struct {
	Status    string   `json:"status"`
	Error     string   `json:"error"`
	RplAmount *big.Int `json:"rplAmount"`
}
type NodeClaimRplResponse struct {
	Status string                `json:"status"`
	Error  string                `json:"error"`
	TxInfo *core.TransactionInfo `json:"txInfo"`
}

type NodeRewardsResponse struct {
	Status                      string                `json:"status"`
	Error                       string                `json:"error"`
	NodeRegistrationTime        time.Time             `json:"nodeRegistrationTime"`
	RewardsInterval             time.Duration         `json:"rewardsInterval"`
	LastCheckpoint              time.Time             `json:"lastCheckpoint"`
	Trusted                     bool                  `json:"trusted"`
	Registered                  bool                  `json:"registered"`
	EffectiveRplStake           float64               `json:"effectiveRplStake"`
	TotalRplStake               float64               `json:"totalRplStake"`
	TrustedRplBond              float64               `json:"trustedRplBond"`
	EstimatedRewards            float64               `json:"estimatedRewards"`
	CumulativeRplRewards        float64               `json:"cumulativeRplRewards"`
	CumulativeEthRewards        float64               `json:"cumulativeEthRewards"`
	EstimatedTrustedRplRewards  float64               `json:"estimatedTrustedRplRewards"`
	CumulativeTrustedRplRewards float64               `json:"cumulativeTrustedRplRewards"`
	UnclaimedRplRewards         float64               `json:"unclaimedRplRewards"`
	UnclaimedEthRewards         float64               `json:"unclaimedEthRewards"`
	UnclaimedTrustedRplRewards  float64               `json:"unclaimedTrustedRplRewards"`
	BeaconRewards               float64               `json:"beaconRewards"`
	TxInfo                      *core.TransactionInfo `json:"txInfo"`
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
	Status string `json:"status"`
	Error  string `json:"error"`
}

type SetSnapshotDelegateResponse struct {
	Status string                `json:"status"`
	Error  string                `json:"error"`
	TxInfo *core.TransactionInfo `json:"txInfo"`
}

type EstimateClearSnapshotDelegateGasResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

type ClearSnapshotDelegateResponse struct {
	Status string                `json:"status"`
	Error  string                `json:"error"`
	TxInfo *core.TransactionInfo `json:"txInfo"`
}

type NodeInitializeFeeDistributorResponse struct {
	Status        string                `json:"status"`
	Error         string                `json:"error"`
	IsInitialized bool                  `json:"isInitialized"`
	Distributor   common.Address        `json:"distributor"`
	TxInfo        *core.TransactionInfo `json:"txInfo"`
}
type NodeDistributeResponse struct {
	Status        string                `json:"status"`
	Error         string                `json:"error"`
	IsInitialized bool                  `json:"isInitialized"`
	Balance       *big.Int              `json:"balance"`
	NodeShare     *big.Int              `json:"nodeShare"`
	TxInfo        *core.TransactionInfo `json:"txInfo"`
}

type NodeGetRewardsInfoResponse struct {
	Status                  string                 `json:"status"`
	Error                   string                 `json:"error"`
	ClaimedIntervals        []uint64               `json:"claimedIntervals"`
	UnclaimedIntervals      []rewards.IntervalInfo `json:"unclaimedIntervals"`
	InvalidIntervals        []rewards.IntervalInfo `json:"invalidIntervals"`
	RplStake                *big.Int               `json:"rplStake"`
	RplPrice                *big.Int               `json:"rplPrice"`
	ActiveMinipools         uint64                 `json:"activeMinipools"`
	EffectiveRplStake       *big.Int               `json:"effectiveRplStake"`
	MinimumRplStake         *big.Int               `json:"minimumRplStake"`
	MaximumRplStake         *big.Int               `json:"maximumRplStake"`
	EthMatched              *big.Int               `json:"ethMatched"`
	EthMatchedLimit         *big.Int               `json:"ethMatchedLimit"`
	PendingMatchAmount      *big.Int               `json:"pendingMatchAmount"`
	BorrowedCollateralRatio float64                `json:"borrowedCollateralRatio"`
	BondedCollateralRatio   float64                `json:"bondedCollateralRatio"`
}

type GetSmoothingPoolRegistrationStatusResponse struct {
	Status                  string        `json:"status"`
	Error                   string        `json:"error"`
	NodeRegistered          bool          `json:"nodeRegistered"`
	TimeLeftUntilChangeable time.Duration `json:"timeLeftUntilChangeable"`
}
type CanSetSmoothingPoolRegistrationStatusResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}
type SetSmoothingPoolRegistrationStatusResponse struct {
	Status string                `json:"status"`
	Error  string                `json:"error"`
	TxInfo *core.TransactionInfo `json:"txInfo"`
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

type NodeCheckCollateralResponse struct {
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
