package api

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/types"
)

type SecurityStatusData struct {
	IsMember       bool   `json:"isMember"`
	CanJoin        bool   `json:"canJoin"`
	CanLeave       bool   `json:"canLeave"`
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

type SecurityMemberDetails struct {
	Address     common.Address `json:"address"`
	Exists      bool           `json:"exists"`
	ID          string         `json:"id"`
	InvitedTime time.Time      `json:"invitedTime"`
	JoinedTime  time.Time      `json:"joinedTime"`
	LeftTime    time.Time      `json:"leftTime"`
}
type SecurityMembersData struct {
	Members []SecurityMemberDetails `json:"members"`
}

type SecurityProposalDetails struct {
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
type SecurityProposalsData struct {
	Proposals []SecurityProposalDetails `json:"proposals"`
}

type SecurityProposeSettingData struct {
	CanPropose     bool                 `json:"canPropose"`
	UnknownSetting bool                 `json:"unknownSetting"`
	TxInfo         *eth.TransactionInfo `json:"txInfo"`
}

type SecurityCancelProposalData struct {
	CanCancel            bool                 `json:"canCancel"`
	DoesNotExist         bool                 `json:"doesNotExist"`
	InvalidState         bool                 `json:"invalidState"`
	InvalidProposer      bool                 `json:"invalidProposer"`
	NotOnSecurityCouncil bool                 `json:"notOnSecurityCouncil"`
	TxInfo               *eth.TransactionInfo `json:"txInfo"`
}

type SecurityVoteOnProposalData struct {
	CanVote            bool                 `json:"canVote"`
	DoesNotExist       bool                 `json:"doesNotExist"`
	InvalidState       bool                 `json:"invalidState"`
	JoinedAfterCreated bool                 `json:"joinedAfterCreated"`
	AlreadyVoted       bool                 `json:"alreadyVoted"`
	TxInfo             *eth.TransactionInfo `json:"txInfo"`
}

type SecurityExecuteProposalData struct {
	CanExecute   bool                 `json:"canExecute"`
	DoesNotExist bool                 `json:"doesNotExist"`
	InvalidState bool                 `json:"invalidState"`
	TxInfo       *eth.TransactionInfo `json:"txInfo"`
}

type SecurityJoinData struct {
	CanJoin         bool                 `json:"canJoin"`
	ProposalExpired bool                 `json:"proposalExpired"`
	AlreadyMember   bool                 `json:"alreadyMember"`
	TxInfo          *eth.TransactionInfo `json:"txInfo"`
}

type SecurityLeaveData struct {
	CanLeave        bool                 `json:"canLeave"`
	IsNotMember     bool                 `json:"isNotMember"`
	ProposalExpired bool                 `json:"proposalExpired"`
	TxInfo          *eth.TransactionInfo `json:"txInfo"`
}
