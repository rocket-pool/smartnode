package voting

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	sharedtypes "github.com/rocket-pool/smartnode/v2/shared/types"
)

const (
	// Request path
	snapshotRequestPath string = "https://%s/graphql?operationName=%s&query=%s"

	// Voting power query
	votingPowerRequest string = `query Vp{
		vp(
			space: "%s",
			voter: "%s",
		) {
			vp
		}
	}`

	// Votes query
	votesRequest string = `query Votes{
		votes(
		  where: {
			space: "%s",
			voter_in: ["%s", "%s"],
		  },
		  orderBy: "created",
		  orderDirection: desc
		) {
		  choice
		  voter
		  proposal {id, state}
		}
	  }`

	// Proposals query
	proposalsRequest string = `query Proposals {
		proposals(where: {space: "%s"%s}, orderBy: "created", orderDirection: desc) {
			title
			choices
			start
			end
			snapshot
			state
			author
			scores
			scores_total
			scores_updated
			quorum
			link
			}
		}`
)

// ===============
// === Structs ===
// ===============

type snapshotVotingPowerResponse struct {
	Data struct {
		Vp struct {
			Vp float64 `json:"vp"`
		} `json:"vp"`
	} `json:"data"`
}

type snapshotProposalVote struct {
	Choice   interface{}    `json:"choice"`
	Voter    common.Address `json:"voter"`
	Proposal struct {
		Id    string `json:"id"`
		State string `json:"state"`
	} `json:"proposal"`
}
type snapshotVotedProposalsResponse struct {
	Data struct {
		Votes []snapshotProposalVote `json:"votes"`
	} `json:"data"`
}

type snapshotProposal struct {
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
type snapshotProposalsResponse struct {
	Data struct {
		Proposals []snapshotProposal `json:"proposals"`
	} `json:"data"`
}

// =============
// === Utils ===
// =============

func GetSnapshotVotingPower(cfg *config.SmartNodeConfig, address common.Address) (float64, error) {
	client := getHttpClientWithTimeout()
	resources := cfg.GetRocketPoolResources()
	apiDomain := resources.SnapshotApiDomain
	id := config.SnapshotID
	if apiDomain == "" {
		return 0, nil
	}

	query := fmt.Sprintf(votingPowerRequest, id, address)
	url := fmt.Sprintf(snapshotRequestPath, apiDomain, "Vp", url.PathEscape(query))
	resp, err := client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("error getting voting power response: %w", err)
	}
	defer resp.Body.Close()

	// Check the response code
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("voting power request failed with code %d", resp.StatusCode)
	}

	// Get response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("error reading voting power response: %w", err)
	}
	var votingPower snapshotVotingPowerResponse
	if err := json.Unmarshal(body, &votingPower); err != nil {
		return 0, fmt.Errorf("error decoding voting power response: %w", err)
	}

	vp := votingPower.Data.Vp.Vp
	return vp, nil
}

func GetSnapshotProposals(cfg *config.SmartNodeConfig, address common.Address, delegate common.Address, activeOnly bool) ([]*sharedtypes.SnapshotProposal, error) {
	client := getHttpClientWithTimeout()
	resources := cfg.GetRocketPoolResources()
	apiDomain := resources.SnapshotApiDomain
	id := config.SnapshotID
	if apiDomain == "" {
		return nil, nil
	}

	// Get the proposals
	snapshotResponse, err := getProposalsResponse(apiDomain, id, client, activeOnly)
	if err != nil {
		return nil, err
	}

	// Get the vote status
	votesResponse, err := getVotesResponse(apiDomain, id, client, address, delegate)
	if err != nil {
		return nil, err
	}

	// Create bindings
	propMap := map[string]*sharedtypes.SnapshotProposal{}
	props := make([]*sharedtypes.SnapshotProposal, len(propMap))
	for i, rawProp := range snapshotResponse.Data.Proposals {
		newProp := &sharedtypes.SnapshotProposal{
			Title:   rawProp.Title,
			State:   sharedtypes.ProposalState(rawProp.State),
			Choices: rawProp.Choices,
			Scores:  rawProp.Scores,
			Quorum:  rawProp.Quorum,
			Link:    rawProp.Link,
			Start:   time.Unix(rawProp.Start, 0),
			End:     time.Unix(rawProp.End, 0),
		}
		props[i] = newProp
		propMap[rawProp.Id] = newProp
	}

	// Map votes to the props
	for _, vote := range votesResponse.Data.Votes {
		prop, exists := propMap[vote.Proposal.Id]
		if !exists {
			continue
		}

		// Get the slice based on whether this is the user or delegate; if it's neither it gets ignored
		var voteSlice *[]int
		if vote.Voter == address {
			voteSlice = &prop.UserVotes
		} else if vote.Voter == delegate {
			voteSlice = &prop.DelegateVotes
		} else {
			continue
		}

		// Add the voter's choices to the prop
		choices, ok := vote.Choice.([]int)
		if ok {
			copy(*voteSlice, choices)
		} else {
			choice, ok := vote.Choice.(int)
			if ok {
				*voteSlice = append(*voteSlice, choice)
			} else {
				*voteSlice = append(*voteSlice, -1)
			}
		}
	}

	return props, nil
}

func getHttpClientWithTimeout() *http.Client {
	return &http.Client{
		Timeout: time.Second * 5,
	}
}

func getProposalsResponse(apiDomain string, id string, client *http.Client, activeOnly bool) (*snapshotProposalsResponse, error) {
	stateFilter := ""
	if activeOnly {
		stateFilter = `, state: "active"`
	}
	query := fmt.Sprintf(proposalsRequest, id, stateFilter)
	url := fmt.Sprintf(snapshotRequestPath, apiDomain, "Proposals", url.PathEscape(query))
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error getting proposals response: %w", err)
	}
	defer resp.Body.Close()

	// Check proposals response code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("proposals request failed with code %d", resp.StatusCode)
	}

	// Get proposals response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var snapshotResponse snapshotProposalsResponse
	if err := json.Unmarshal(body, &snapshotResponse); err != nil {
		return nil, fmt.Errorf("error decoding proposals response: %w", err)
	}
	return &snapshotResponse, nil
}

func getVotesResponse(apiDomain string, id string, client *http.Client, address common.Address, delegate common.Address) (*snapshotVotedProposalsResponse, error) {
	query := fmt.Sprintf(votesRequest, id, address, delegate)
	url := fmt.Sprintf(snapshotRequestPath, apiDomain, "Votes", url.PathEscape(query))
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error getting vote count response: %w", err)
	}
	defer resp.Body.Close()

	// Check the response code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vote count request failed with code %d", resp.StatusCode)
	}

	// Get response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro reading vote count response: %w", err)
	}
	var votedProposals snapshotVotedProposalsResponse
	if err := json.Unmarshal(body, &votedProposals); err != nil {
		return nil, fmt.Errorf("error decoding vote count response: %w", err)
	}
	return &votedProposals, nil
}
