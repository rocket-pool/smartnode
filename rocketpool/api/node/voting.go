package node

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getHttpClientWithTimeout() *http.Client {
	return &http.Client{
		Timeout: time.Second * 5,
	}
}

func GetSnapshotVotingPower(apiDomain string, space string, nodeAddress common.Address) (*api.SnapshotVotingPower, error) {
	client := getHttpClientWithTimeout()
	query := fmt.Sprintf(`query Vp{
		vp(
			space: "%s",
			voter: "%s",
		) {
			vp
		}
	}
	`, space, nodeAddress)
	url := fmt.Sprintf("https://%s/graphql?operationName=Vp&query=%s", apiDomain, url.PathEscape(query))
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// Check the response code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with code %d", resp.StatusCode)
	}

	// Get response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var votingPower api.SnapshotVotingPower
	if err := json.Unmarshal(body, &votingPower); err != nil {
		return nil, fmt.Errorf("could not decode snapshot response: %w", err)

	}

	return &votingPower, nil
}

func GetSnapshotVotedProposals(apiDomain string, space string, nodeAddress common.Address, delegate common.Address) (*api.NodeSnapshotVotedProposalsData, error) {
	client := getHttpClientWithTimeout()
	query := fmt.Sprintf(`query Votes{
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
	  }`, space, nodeAddress, delegate)
	url := fmt.Sprintf("https://%s/graphql?operationName=Votes&query=%s", apiDomain, url.PathEscape(query))
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// Check the response code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with code %d", resp.StatusCode)
	}

	// Get response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var votedProposals api.NodeSnapshotVotedProposalsData
	if err := json.Unmarshal(body, &votedProposals); err != nil {
		return nil, fmt.Errorf("could not decode snapshot response: %w", err)

	}

	return &votedProposals, nil
}

func GetSnapshotProposals(apiDomain string, space string, state string) (*api.NodeSnapshotData, error) {
	client := getHttpClientWithTimeout()
	stateFilter := ""
	if state != "" {
		stateFilter = fmt.Sprintf(`, state: "%s"`, state)
	}
	query := fmt.Sprintf(`query Proposals {
	proposals(where: {space: "%s"%s}, orderBy: "created", orderDirection: desc) {
	    id
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
    }`, space, stateFilter)

	url := fmt.Sprintf("https://%s/graphql?operationName=Proposals&query=%s", apiDomain, url.PathEscape(query))
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// Check the response code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with code %d", resp.StatusCode)
	}

	// Get response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var snapshotResponse api.NodeSnapshotData
	if err := json.Unmarshal(body, &snapshotResponse); err != nil {
		return nil, fmt.Errorf("Could not decode snapshot response: %w", err)

	}

	return &snapshotResponse, nil
}
