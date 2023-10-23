package api

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/core"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	sharedtypes "github.com/rocket-pool/smartnode/shared/types"
	"github.com/rocket-pool/smartnode/shared/utils/rp"
)

type NodeStatusData struct {
	AccountAddress                    common.Address `json:"accountAddress"`
	AccountAddressFormatted           string         `json:"accountAddressFormatted"`
	WithdrawalAddress                 common.Address `json:"withdrawalAddress"`
	WithdrawalAddressFormatted        string         `json:"withdrawalAddressFormatted"`
	PendingWithdrawalAddress          common.Address `json:"pendingWithdrawalAddress"`
	PendingWithdrawalAddressFormatted string         `json:"pendingWithdrawalAddressFormatted"`
	Registered                        bool           `json:"registered"`
	Trusted                           bool           `json:"trusted"`
	TimezoneLocation                  string         `json:"timezoneLocation"`
	NodeBalances                      struct {
		Eth   *big.Int `json:"eth"`
		Reth  *big.Int `json:"reth"`
		Rpl   *big.Int `json:"rpl"`
		Fsrpl *big.Int `json:"fsrpl"`
	} `json:"nodeBalances"`
	WithdrawalBalances struct {
		Eth   *big.Int `json:"eth"`
		Reth  *big.Int `json:"reth"`
		Rpl   *big.Int `json:"rpl"`
		Fsrpl *big.Int `json:"fsrpl"`
	} `json:"withdrawalBalances"`
	RplStake                       *big.Int       `json:"rplStake"`
	EffectiveRplStake              *big.Int       `json:"effectiveRplStake"`
	MinimumRplStake                *big.Int       `json:"minimumRplStake"`
	MaximumRplStake                *big.Int       `json:"maximumRplStake"`
	BorrowedCollateralRatio        float64        `json:"borrowedCollateralRatio"`
	BondedCollateralRatio          float64        `json:"bondedCollateralRatio"`
	PendingEffectiveRplStake       *big.Int       `json:"pendingEffectiveRplStake"`
	PendingMinimumRplStake         *big.Int       `json:"pendingMinimumRplStake"`
	PendingMaximumRplStake         *big.Int       `json:"pendingMaximumRplStake"`
	PendingBorrowedCollateralRatio float64        `json:"pendingBorrowedCollateralRatio"`
	PendingBondedCollateralRatio   float64        `json:"pendingBondedCollateralRatio"`
	VotingDelegate                 common.Address `json:"votingDelegate"`
	VotingDelegateFormatted        string         `json:"votingDelegateFormatted"`
	MinipoolLimit                  uint64         `json:"minipoolLimit"`
	EthMatched                     *big.Int       `json:"ethMatched"`
	EthMatchedLimit                *big.Int       `json:"ethMatchedLimit"`
	PendingMatchAmount             *big.Int       `json:"pendingMatchAmount"`
	CreditBalance                  *big.Int       `json:"creditBalance"`
	MinipoolCounts                 struct {
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

type NodeRegisterData struct {
	CanRegister          bool                  `json:"canRegister"`
	AlreadyRegistered    bool                  `json:"alreadyRegistered"`
	RegistrationDisabled bool                  `json:"registrationDisabled"`
	TxInfo               *core.TransactionInfo `json:"txInfo"`
}

type NodeSetWithdrawalAddressData struct {
	CanSet bool                  ` json:"canSet"`
	TxInfo *core.TransactionInfo `json:"txInfo"`
}

type NodeConfirmWithdrawalAddressData struct {
	CanConfirm bool                  `json:"canConfirm"`
	TxInfo     *core.TransactionInfo `json:"txInfo"`
}

type NodeGetWithdrawalAddressData struct {
	Address common.Address `json:"address"`
}

type NodeGetPendingWithdrawalAddressData struct {
	Address common.Address `json:"address"`
}

type NodeSetTimezoneData struct {
	CanSet bool                  `json:"canSet"`
	TxInfo *core.TransactionInfo `json:"txInfo"`
}

type NodeSwapRplData struct {
	CanSwap             bool                  `json:"canSwap"`
	InsufficientBalance bool                  `json:"insufficientBalance"`
	Allowance           *big.Int              `json:"allowance"`
	ApproveTxInfo       *core.TransactionInfo `json:"approveTxInfo"`
	SwapTxInfo          *core.TransactionInfo `json:"swapTxInfo"`
}

type NodeStakeRplData struct {
	CanStake            bool                  `json:"canStake"`
	InsufficientBalance bool                  `json:"insufficientBalance"`
	Allowance           *big.Int              `json:"allowance"`
	ApproveTxInfo       *core.TransactionInfo `json:"approveTxInfo"`
	StakeTxInfo         *core.TransactionInfo `json:"stakeTxInfo"`
}

type NodeSetStakeRplForAllowedData struct {
	CanSet bool                  `json:"canSet"`
	TxInfo *core.TransactionInfo `json:"txInfo"`
}

type NodeWithdrawRplData struct {
	CanWithdraw                  bool                  `json:"canWithdraw"`
	InsufficientBalance          bool                  `json:"insufficientBalance"`
	MinipoolsUndercollateralized bool                  `json:"minipoolsUndercollateralized"`
	WithdrawalDelayActive        bool                  `json:"withdrawalDelayActive"`
	TxInfo                       *core.TransactionInfo `json:"txInfo"`
}

type NodeDepositData struct {
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

type NodeCreateVacantMinipoolData struct {
	CanDeposit            bool                  `json:"canDeposit"`
	InsufficientRplStake  bool                  `json:"insufficientRplStake"`
	InvalidAmount         bool                  `json:"invalidAmount"`
	DepositDisabled       bool                  `json:"depositDisabled"`
	MinipoolAddress       common.Address        `json:"minipoolAddress"`
	ScrubPeriod           time.Duration         `json:"scrubPeriod"`
	WithdrawalCredentials common.Hash           `json:"withdrawalCredentials"`
	TxInfo                *core.TransactionInfo `json:"txInfo"`
}

type NodeSendData struct {
	Balance             *big.Int              `json:"balance"`
	TokenName           string                `json:"name"`
	TokenSymbol         string                `json:"symbol"`
	CanSend             bool                  `json:"canSend"`
	InsufficientBalance bool                  `json:"insufficientBalance"`
	TxInfo              *core.TransactionInfo `json:"txInfo"`
}

type NodeSendMessageData struct {
	TxInfo *core.TransactionInfo `json:"txInfo"`
}

type NodeBurnData struct {
	CanBurn                bool                  `json:"canBurn"`
	InsufficientBalance    bool                  `json:"insufficientBalance"`
	InsufficientCollateral bool                  `json:"insufficientCollateral"`
	TxInfo                 *core.TransactionInfo `json:"txInfo"`
}

type NodeSyncProgressData struct {
	EcStatus ClientManagerStatus `json:"ecStatus"`
	BcStatus ClientManagerStatus `json:"bcStatus"`
}

type NodeRewardsData struct {
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

type NodeSignData struct {
	SignedData string `json:"signedData"`
}

type NodeInitializeFeeDistributorData struct {
	CanInitialize bool                  `json:"canInitialize"`
	IsInitialized bool                  `json:"isInitialized"`
	Distributor   common.Address        `json:"distributor"`
	TxInfo        *core.TransactionInfo `json:"txInfo"`
}

type NodeDistributeData struct {
	CanDistribute bool                  `json:"canDistribute"`
	NoBalance     bool                  `json:"noBalance"`
	IsInitialized bool                  `json:"isInitialized"`
	Balance       *big.Int              `json:"balance"`
	NodeShare     *big.Int              `json:"nodeShare"`
	TxInfo        *core.TransactionInfo `json:"txInfo"`
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
	MaximumRplStake         *big.Int                   `json:"maximumRplStake"`
	EthMatched              *big.Int                   `json:"ethMatched"`
	EthMatchedLimit         *big.Int                   `json:"ethMatchedLimit"`
	PendingMatchAmount      *big.Int                   `json:"pendingMatchAmount"`
	BorrowedCollateralRatio float64                    `json:"borrowedCollateralRatio"`
	BondedCollateralRatio   float64                    `json:"bondedCollateralRatio"`
}

type NodeSetSmoothingPoolRegistrationStatusData struct {
	NodeRegistered          bool                  `json:"nodeRegistered"`
	CanChange               bool                  `json:"canChange"`
	TimeLeftUntilChangeable time.Duration         `json:"timeLeftUntilChangeable"`
	TxInfo                  *core.TransactionInfo `json:"txInfo"`
}

type NodeResolveEnsData struct {
	Address       common.Address `json:"address"`
	EnsName       string         `json:"ensName"`
	FormattedName string         `json:"formattedName"`
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
type NodeSnapshotData struct {
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
type NodeSnapshotVotedProposalsData struct {
	Data struct {
		Votes []SnapshotProposalVote `json:"votes"`
	} `json:"data"`
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
