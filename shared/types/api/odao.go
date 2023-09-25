package api

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/types"
)

var oracleDaoSettingNamesSingleton *OracleDaoSettingNames

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
	CanPropose             bool                  `json:"canPropose"`
	ProposalCooldownActive bool                  `json:"proposalCooldownActive"`
	MemberAlreadyExists    bool                  `json:"memberAlreadyExists"`
	TxInfo                 *core.TransactionInfo `json:"txInfo"`
}

type OracleDaoProposeLeaveData struct {
	CanPropose             bool                  `json:"canPropose"`
	ProposalCooldownActive bool                  `json:"proposalCooldownActive"`
	InsufficientMembers    bool                  `json:"insufficientMembers"`
	TxInfo                 *core.TransactionInfo `json:"txInfo"`
}

type OracleDaoProposeReplaceData struct {
	CanPropose             bool                  `json:"canPropose"`
	ProposalCooldownActive bool                  `json:"proposalCooldownActive"`
	MemberAlreadyExists    bool                  `json:"memberAlreadyExists"`
	TxInfo                 *core.TransactionInfo `json:"txInfo"`
}

type OracleDaoProposeKickData struct {
	Status                 string                `json:"status"`
	Error                  string                `json:"error"`
	CanPropose             bool                  `json:"canPropose"`
	MemberDoesNotExist     bool                  `json:"memberDoesNotExist"`
	ProposalCooldownActive bool                  `json:"proposalCooldownActive"`
	InsufficientRplBond    bool                  `json:"insufficientRplBond"`
	TxInfo                 *core.TransactionInfo `json:"txInfo"`
}

type OracleDaoCancelProposalData struct {
	CanCancel       bool                  `json:"canCancel"`
	DoesNotExist    bool                  `json:"doesNotExist"`
	InvalidState    bool                  `json:"invalidState"`
	InvalidProposer bool                  `json:"invalidProposer"`
	TxInfo          *core.TransactionInfo `json:"txInfo"`
}

type OracleDaoVoteData struct {
	CanVote            bool                  `json:"canVote"`
	DoesNotExist       bool                  `json:"doesNotExist"`
	InvalidState       bool                  `json:"invalidState"`
	JoinedAfterCreated bool                  `json:"joinedAfterCreated"`
	AlreadyVoted       bool                  `json:"alreadyVoted"`
	TxInfo             *core.TransactionInfo `json:"txInfo"`
}

type OracleDaoExecuteProposalData struct {
	CanExecute   bool                  `json:"canExecute"`
	DoesNotExist bool                  `json:"doesNotExist"`
	InvalidState bool                  `json:"invalidState"`
	TxInfo       *core.TransactionInfo `json:"txInfo"`
}

type OracleDaoJoinData struct {
	CanJoin                bool                  `json:"canJoin"`
	ProposalExpired        bool                  `json:"proposalExpired"`
	AlreadyMember          bool                  `json:"alreadyMember"`
	InsufficientRplBalance bool                  `json:"insufficientRplBalance"`
	ApproveTxInfo          *core.TransactionInfo `json:"approveTxInfo"`
	JoinTxInfo             *core.TransactionInfo `json:"joinTxInfo"`
}

type OracleDaoLeaveData struct {
	CanLeave            bool                  `json:"canLeave"`
	ProposalExpired     bool                  `json:"proposalExpired"`
	InsufficientMembers bool                  `json:"insufficientMembers"`
	TxInfo              *core.TransactionInfo `json:"txInfo"`
}

type OracleDaoReplaceData struct {
	CanReplace          bool                  `json:"canReplace"`
	ProposalExpired     bool                  `json:"proposalExpired"`
	MemberAlreadyExists bool                  `json:"memberAlreadyExists"`
	TxInfo              *core.TransactionInfo `json:"txInfo"`
}

type OracleDaoProposeSettingData struct {
	CanPropose             bool                  `json:"canPropose"`
	ProposalCooldownActive bool                  `json:"proposalCooldownActive"`
	TxInfo                 *core.TransactionInfo `json:"txInfo"`
}

type OracleDaoSettingsData struct {
	Member struct {
		Quorum              float64       `json:"quorum"`
		RplBond             *big.Int      `json:"rplBond"`
		MinipoolUnbondedMax uint64        `json:"minipoolUnbondedMax"`
		ChallengeCooldown   time.Duration `json:"challengeCooldown"`
		ChallengeWindow     time.Duration `json:"challengeWindow"`
		ChallengeCost       *big.Int      `json:"challengeCost"`
	} `json:"member"`

	Minipool struct {
		ScrubPeriod               time.Duration `json:"scrubPeriod"`
		PromotionScrubPeriod      time.Duration `json:"promotionScrubPeriod"`
		ScrubPenaltyEnabled       bool          `json:"scrubPenaltyEnabled"`
		BondReductionWindowStart  time.Duration `json:"bondReductionWindowStart"`
		BondReductionWindowLength time.Duration `json:"bondReductionWindowLength"`
	} `json:"minipool"`

	Proposal struct {
		Cooldown      time.Duration `json:"cooldown"`
		VoteTime      time.Duration `json:"voteTime"`
		VoteDelayTime time.Duration `json:"voteDelayTime"`
		ExecuteTime   time.Duration `json:"executeTime"`
		ActionTime    time.Duration `json:"actionTime"`
	} `json:"proposal"`
}

type OracleDaoSettingNames struct {
	Member struct {
		Quorum              string
		RplBond             string
		MinipoolUnbondedMax string
		ChallengeCooldown   string
		ChallengeWindow     string
		ChallengeCost       string
	}

	Minipool struct {
		ScrubPeriod               string
		PromotionScrubPeriod      string
		ScrubPenaltyEnabled       string
		BondReductionWindowStart  string
		BondReductionWindowLength string
	}

	Proposal struct {
		Cooldown      string
		VoteTime      string
		VoteDelayTime string
		ExecuteTime   string
		ActionTime    string
	}
}

func GetOracleDaoSettingsNames() *OracleDaoSettingNames {
	// Return the singleton
	if oracleDaoSettingNamesSingleton != nil {
		return oracleDaoSettingNamesSingleton
	}
	names := &OracleDaoSettingNames{}

	// Member
	names.Member.Quorum = "member.quorum"
	names.Member.RplBond = "member.rplBond"
	names.Member.MinipoolUnbondedMax = "member.minipoolUnbondedMax"
	names.Member.ChallengeCooldown = "member.challengeCooldown"
	names.Member.ChallengeWindow = "member.challengeWindow"
	names.Member.ChallengeCost = "member.challengeCost"

	// Minipool
	names.Minipool.ScrubPeriod = "minipool.scrubPeriod"
	names.Minipool.PromotionScrubPeriod = "minipool.promotionScrubPeriod"
	names.Minipool.ScrubPenaltyEnabled = "minipool.scrubPenaltyEnabled"
	names.Minipool.BondReductionWindowStart = "minipool.bondReductionWindowStart"
	names.Minipool.BondReductionWindowLength = "minipool.bondReductionWindowLength"

	// Proposal
	names.Proposal.Cooldown = "proposal.cooldown"
	names.Proposal.VoteTime = "proposal.voteTime"
	names.Proposal.VoteDelayTime = "proposal.voteDelayTime"
	names.Proposal.ExecuteTime = "proposal.executeTime"
	names.Proposal.ActionTime = "proposal.actionTime"

	oracleDaoSettingNamesSingleton = names
	return names
}
