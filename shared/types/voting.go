package types

import "time"

type ProposalState string

const (
	ProposalState_Active  ProposalState = "active"
	ProposalState_Pending ProposalState = "pending"
	ProposalState_Closed  ProposalState = "closed"
)

type SnapshotProposal struct {
	Title         string        `json:"title"`
	State         ProposalState `json:"state"`
	Start         time.Time     `json:"start"`
	End           time.Time     `json:"end"`
	Choices       []string      `json:"choices"`
	Scores        []float64     `json:"scores"`
	Quorum        float64       `json:"quorum"`
	Link          string        `json:"link"`
	UserVotes     []int         `json:"userVotes"`
	DelegateVotes []int         `json:"delegateVotes"`
}
