package api

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/types"
)

type OracleDaoStatusData struct {
	IsMember       bool   `json:"isMember"`
	CanJoin        bool   `json:"canJoin"`
	CanLeave       bool   `json:"canLeave"`
	CanReplace     bool   `json:"canReplace"`
	TotalMembers   uint64 `json:"totalMembers"`
	ProposalCounts struct {
		Total     int `json:"total"`
		Pending   int `json:"pending"`
		Active    int `json:"active"`
		Cancelled int `json:"cancelled"`
		Defeated  int `json:"defeated"`
		Succeeded int `json:"succeeded"`
		Expired   int `json:"expired"`
		Executed  int `json:"executed"`
	} `json:"proposalCounts"`
}

type OracleDaoMemberDetails struct {
	Address          common.Address `json:"address"`
	Exists           bool           `json:"exists"`
	ID               string         `json:"id"`
	Url              string         `json:"url"`
	JoinedTime       time.Time      `json:"joinedTime"`
	LastProposalTime time.Time      `json:"lastProposalTime"`
	RplBondAmount    *big.Int       `json:"rplBondAmount"`
}
type OracleDaoMembersData struct {
	Members []OracleDaoMemberDetails `json:"members"`
}

type OracleDaoProposalDetails struct {
	ID              uint64              `json:"id"`
	ProposerAddress common.Address      `json:"proposerAddress"`
	Message         string              `json:"message"`
	CreatedTime     time.Time           `json:"createdTime"`
	StartTime       time.Time           `json:"startTime"`
	EndTime         time.Time           `json:"endTime"`
	ExpiryTime      time.Time           `json:"expiryTime"`
	VotesRequired   float64             `json:"votesRequired"`
	VotesFor        float64             `json:"votesFor"`
	VotesAgainst    float64             `json:"votesAgainst"`
	MemberVoted     bool                `json:"memberVoted"`
	MemberSupported bool                `json:"memberSupported"`
	IsCancelled     bool                `json:"isCancelled"`
	IsExecuted      bool                `json:"isExecuted"`
	Payload         []byte              `json:"payload"`
	PayloadStr      string              `json:"payloadStr"`
	State           types.ProposalState `json:"state"`
}
type OracleDaoProposalsData struct {
	Proposals []OracleDaoProposalDetails `json:"proposals"`
}

type OracleDaoProposeInviteData struct {
	CanPropose             bool                 `json:"canPropose"`
	ProposalCooldownActive bool                 `json:"proposalCooldownActive"`
	MemberAlreadyExists    bool                 `json:"memberAlreadyExists"`
	TxInfo                 *eth.TransactionInfo `json:"txInfo"`
}

type OracleDaoProposeLeaveData struct {
	CanPropose             bool                 `json:"canPropose"`
	ProposalCooldownActive bool                 `json:"proposalCooldownActive"`
	MemberDoesntExist      bool                 `json:"memberDoesntExist"`
	InsufficientMembers    bool                 `json:"insufficientMembers"`
	TxInfo                 *eth.TransactionInfo `json:"txInfo"`
}

type OracleDaoProposeKickData struct {
	Status                 string               `json:"status"`
	Error                  string               `json:"error"`
	CanPropose             bool                 `json:"canPropose"`
	MemberDoesNotExist     bool                 `json:"memberDoesNotExist"`
	ProposalCooldownActive bool                 `json:"proposalCooldownActive"`
	InsufficientRplBond    bool                 `json:"insufficientRplBond"`
	TxInfo                 *eth.TransactionInfo `json:"txInfo"`
}

type OracleDaoCancelProposalData struct {
	CanCancel       bool                 `json:"canCancel"`
	DoesNotExist    bool                 `json:"doesNotExist"`
	InvalidState    bool                 `json:"invalidState"`
	InvalidProposer bool                 `json:"invalidProposer"`
	TxInfo          *eth.TransactionInfo `json:"txInfo"`
}

type OracleDaoVoteOnProposalData struct {
	CanVote            bool                 `json:"canVote"`
	DoesNotExist       bool                 `json:"doesNotExist"`
	InvalidState       bool                 `json:"invalidState"`
	JoinedAfterCreated bool                 `json:"joinedAfterCreated"`
	AlreadyVoted       bool                 `json:"alreadyVoted"`
	TxInfo             *eth.TransactionInfo `json:"txInfo"`
}

type OracleDaoExecuteProposalData struct {
	CanExecute   bool                 `json:"canExecute"`
	DoesNotExist bool                 `json:"doesNotExist"`
	InvalidState bool                 `json:"invalidState"`
	TxInfo       *eth.TransactionInfo `json:"txInfo"`
}

type OracleDaoJoinData struct {
	CanJoin                bool                 `json:"canJoin"`
	ProposalExpired        bool                 `json:"proposalExpired"`
	AlreadyMember          bool                 `json:"alreadyMember"`
	InsufficientRplBalance bool                 `json:"insufficientRplBalance"`
	ApproveTxInfo          *eth.TransactionInfo `json:"approveTxInfo"`
	JoinTxInfo             *eth.TransactionInfo `json:"joinTxInfo"`
}

type OracleDaoLeaveData struct {
	CanLeave            bool                 `json:"canLeave"`
	ProposalExpired     bool                 `json:"proposalExpired"`
	InsufficientMembers bool                 `json:"insufficientMembers"`
	TxInfo              *eth.TransactionInfo `json:"txInfo"`
}

type OracleDaoProposeSettingData struct {
	CanPropose             bool                 `json:"canPropose"`
	UnknownSetting         bool                 `json:"unknownSetting"`
	ProposalCooldownActive bool                 `json:"proposalCooldownActive"`
	TxInfo                 *eth.TransactionInfo `json:"txInfo"`
}

type OracleDaoSettingsData struct {
	Member struct {
		Quorum            float64       `json:"quorum"`
		RplBond           *big.Int      `json:"rplBond"`
		ChallengeCooldown time.Duration `json:"challengeCooldown"`
		ChallengeWindow   time.Duration `json:"challengeWindow"`
		ChallengeCost     *big.Int      `json:"challengeCost"`
	} `json:"member"`

	Minipool struct {
		ScrubPeriod                     time.Duration `json:"scrubPeriod"`
		ScrubQuorum                     float64       `json:"scrubQuorum"`
		PromotionScrubPeriod            time.Duration `json:"promotionScrubPeriod"`
		IsScrubPenaltyEnabled           bool          `json:"isScrubPenaltyEnabled"`
		BondReductionWindowStart        time.Duration `json:"bondReductionWindowStart"`
		BondReductionWindowLength       time.Duration `json:"bondReductionWindowLength"`
		BondReductionCancellationQuorum float64       `json:"bondReductionCancellationQuorum"`
	} `json:"minipool"`

	Proposal struct {
		Cooldown      time.Duration `json:"cooldown"`
		VoteTime      time.Duration `json:"voteTime"`
		VoteDelayTime time.Duration `json:"voteDelayTime"`
		ExecuteTime   time.Duration `json:"executeTime"`
		ActionTime    time.Duration `json:"actionTime"`
	} `json:"proposal"`
}
