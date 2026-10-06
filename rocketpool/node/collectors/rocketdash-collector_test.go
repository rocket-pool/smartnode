package collectors

import (
	"testing"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

func TestRocketDashCounts(t *testing.T) {
	status := api.SnapshotResponseStruct{
		ActiveSnapshotProposals: []api.SnapshotProposal{
			{Id: "active", State: "active"},
			{Id: "unvoted", State: "active"},
			{Id: "closed", State: "closed"},
			{Id: "pending", State: "pending"},
		},
	}
	for _, id := range []string{"active", "active", "closed", "missing", "pending"} {
		vote := api.SnapshotProposalVote{}
		vote.Proposal.Id = id
		status.ProposalVotes = append(status.ProposalVotes, vote)
	}
	collector := &RocketDashCollector{}
	collector.collectProposalsAndVotes(status)
	if collector.cachedActiveProposals != 2 || collector.cachedClosedProposals != 1 || collector.cachedVotesActiveProposals != 1 || collector.cachedVotesClosedProposals != 1 {
		t.Fatalf("wrong proposal/vote counts: %+v", collector)
	}
	collector.collectProposalsAndVotes(api.SnapshotResponseStruct{})
	if collector.cachedActiveProposals != 0 || collector.cachedClosedProposals != 0 || collector.cachedVotesActiveProposals != 0 || collector.cachedVotesClosedProposals != 0 {
		t.Fatal("empty refresh retained old counts")
	}
}
