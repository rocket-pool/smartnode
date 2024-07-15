package types

import "time"

type ProposalState string

const (
	ProposalState_Active  ProposalState = "active"
	ProposalState_Pending ProposalState = "pending"
	ProposalState_Closed  ProposalState = "closed"
)

type SnapshotProposal struct {
	Id            string        `json:"id"`
	Title         string        `json:"title"`
	Start         time.Time     `json:"start"`
	End           time.Time     `json:"end"`
	State         ProposalState `json:"state"`
	Snapshot      string        `json:"snapshot"`
	Author        string        `json:"author"`
	Choices       []string      `json:"choices"`
	Scores        []float64     `json:"scores"`
	ScoresTotal   float64       `json:"score_total"`
	ScoresUpdated int64         `json:"scores_updated"`
	Quorum        float64       `json:"quorum"`
	Link          string        `json:"link"`
	UserVotes     []int         `json:"userVotes"`
	DelegateVotes []int         `json:"delegateVotes"`
}
