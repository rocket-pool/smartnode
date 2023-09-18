package api

import (
	"math/big"

	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/dao"
	tn "github.com/rocket-pool/rocketpool-go/dao/trustednode"
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

type OracleDaoMembersData struct {
	Members []tn.MemberDetails `json:"members"`
}

type OracleDaoProposalsData struct {
	Proposals []dao.ProposalDetails `json:"proposals"`
}

type OracleDaoProposalData struct {
	Proposal dao.ProposalDetails `json:"proposal"`
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
	ProposalCooldownActive bool                  `json:"proposalCooldownActive"`
	InsufficientRplBond    bool                  `json:"insufficientRplBond"`
	TxInfo                 *core.TransactionInfo `json:"txInfo"`
}

type OracleDaoProposalCancelData struct {
	CanCancel       bool                  `json:"canCancel"`
	DoesNotExist    bool                  `json:"doesNotExist"`
	InvalidState    bool                  `json:"invalidState"`
	InvalidProposer bool                  `json:"invalidProposer"`
	TxInfo          *core.TransactionInfo `json:"txInfo"`
}

type OracleDaoProposalVoteData struct {
	CanVote            bool                  `json:"canVote"`
	DoesNotExist       bool                  `json:"doesNotExist"`
	InvalidState       bool                  `json:"invalidState"`
	JoinedAfterCreated bool                  `json:"joinedAfterCreated"`
	AlreadyVoted       bool                  `json:"alreadyVoted"`
	TxInfo             *core.TransactionInfo `json:"txInfo"`
}

type OracleDaoProposalExecuteData struct {
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

type OracleDaoMemberSettingsData struct {
	Quorum              float64  `json:"quorum"`
	RPLBond             *big.Int `json:"rplBond"`
	MinipoolUnbondedMax uint64   `json:"minipoolUnbondedMax"`
	ChallengeCooldown   uint64   `json:"challengeCooldown"`
	ChallengeWindow     uint64   `json:"challengeWindow"`
	ChallengeCost       *big.Int `json:"challengeCost"`
}

type OracleDaoProposalSettingsData struct {
	Cooldown      uint64 `json:"cooldown"`
	VoteTime      uint64 `json:"voteTime"`
	VoteDelayTime uint64 `json:"voteDelayTime"`
	ExecuteTime   uint64 `json:"executeTime"`
	ActionTime    uint64 `json:"actionTime"`
}

type OracleDaoMinipoolSettingsData struct {
	ScrubPeriod               uint64 `json:"scrubPeriod"`
	PromotionScrubPeriod      uint64 `json:"promotionScrubPeriod"`
	ScrubPenaltyEnabled       bool   `json:"scrubPenaltyEnabled"`
	BondReductionWindowStart  uint64 `json:"bondReductionWindowStart"`
	BondReductionWindowLength uint64 `json:"bondReductionWindowLength"`
}
