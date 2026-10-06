package governance

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

const DefaultURL = "https://rocketdash.net"

// Client reads the same public REST API as https://rocketdash.net/vote.
// The Snapshot response types are retained for Smartnode API compatibility.
type Client struct {
	baseURL string
	network string
	http    *http.Client
}

func NewClient(baseURL string, chainID uint) (*Client, error) {
	var network string
	switch chainID {
	case 1:
		network = "mainnet"
	case 560048:
		network = "hoodi"
	default:
		return nil, fmt.Errorf("RocketDash voting is not available for chain %d", chainID)
	}
	if baseURL == "" {
		baseURL = DefaultURL
	}
	return &Client{strings.TrimRight(baseURL, "/"), network, &http.Client{Timeout: 5 * time.Second}}, nil
}

func (c *Client) get(ctx context.Context, path string, query url.Values, result any) error {
	query.Set("network", c.network)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/gov/"+path+"?"+query.Encode(), nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("RocketDash %s: %w", path, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("RocketDash %s returned HTTP %d", path, resp.StatusCode)
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 8<<20)).Decode(result); err != nil {
		return fmt.Errorf("could not decode RocketDash %s: %w", path, err)
	}
	return nil
}

func (c *Client) Proposals(ctx context.Context, state string) ([]api.SnapshotProposal, error) {
	var response struct {
		Network   string `json:"network"`
		Proposals *[]struct {
			ID          string    `json:"proposal_id"`
			Title       string    `json:"title"`
			Start       int64     `json:"start_time"`
			End         int64     `json:"end_time"`
			Snapshot    int64     `json:"snapshot_block"`
			Lifecycle   string    `json:"lifecycle"`
			State       string    `json:"state"`
			Author      string    `json:"author"`
			Choices     []string  `json:"choices"`
			Scores      []float64 `json:"scores"`
			ScoresTotal float64   `json:"scores_total"`
			Quorum      float64   `json:"quorum"`
		} `json:"proposals"`
	}
	if err := c.get(ctx, "proposals", url.Values{}, &response); err != nil {
		return nil, err
	}
	if response.Network != c.network || response.Proposals == nil {
		return nil, fmt.Errorf("invalid RocketDash proposals response for %s", c.network)
	}
	proposals := make([]api.SnapshotProposal, 0, len(*response.Proposals))
	for _, p := range *response.Proposals {
		// A proposal cannot be voted on until its power snapshot is ready.
		if p.Lifecycle != "ready" || (state != "" && p.State != state) {
			continue
		}
		if p.ID == "" || p.Snapshot < 0 || p.Snapshot > int64(^uint32(0)) || len(p.Scores) != len(p.Choices) {
			return nil, fmt.Errorf("invalid RocketDash proposal %q", p.ID)
		}
		link := c.baseURL + "/vote/" + url.PathEscape(p.ID)
		if c.network != "mainnet" {
			link += "?network=" + c.network
		}
		proposals = append(proposals, api.SnapshotProposal{
			Id: p.ID, Title: p.Title, Start: p.Start, End: p.End, State: p.State,
			Snapshot: p.Snapshot, Author: p.Author, Choices: p.Choices, Scores: p.Scores,
			ScoresTotal: p.ScoresTotal, Quorum: p.Quorum, Link: link,
		})
	}
	return proposals, nil
}

// Votes returns at most one effective ballot per proposal. RocketDash publishes
// the node behind each signer, so a changed signalling address does not hide an
// earlier ballot. A node's own ballot overrides its delegate's ballot. Resolve
// delegation at each proposal's snapshot block, rather than at the current head.
func (c *Client) Votes(ctx context.Context, proposals []api.SnapshotProposal, node common.Address, delegateAt func(uint32) (common.Address, error)) ([]api.SnapshotProposalVote, error) {
	results := make([]*api.SnapshotProposalVote, len(proposals))
	group, ctx := errgroup.WithContext(ctx)
	group.SetLimit(4)
	for i, proposal := range proposals {
		if proposal.State != "active" && proposal.State != "closed" {
			continue
		}
		group.Go(func() error {
			var response struct {
				Votes *[]struct {
					Node   common.Address `json:"node"`
					Choice any            `json:"choice"`
				} `json:"votes"`
			}
			if err := c.get(ctx, "proposal-votes", url.Values{"id": {proposal.Id}}, &response); err != nil {
				return err
			}
			if response.Votes == nil {
				return fmt.Errorf("invalid RocketDash votes response for %s", proposal.Id)
			}
			findVote := func(address common.Address) *api.SnapshotProposalVote {
				if address == (common.Address{}) {
					return nil
				}
				for _, v := range *response.Votes {
					if v.Node == address && v.Choice != nil {
						vote := &api.SnapshotProposalVote{Voter: v.Node, Choice: v.Choice}
						vote.Proposal.Id = proposal.Id
						vote.Proposal.State = proposal.State
						return vote
					}
				}
				return nil
			}
			results[i] = findVote(node)
			if results[i] == nil && delegateAt != nil {
				delegate, err := delegateAt(uint32(proposal.Snapshot))
				if err != nil {
					return fmt.Errorf("getting delegate at RocketDash proposal snapshot %d: %w", proposal.Snapshot, err)
				}
				results[i] = findVote(delegate)
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return nil, err
	}
	votes := make([]api.SnapshotProposalVote, 0, len(results))
	for _, vote := range results {
		if vote != nil {
			votes = append(votes, *vote)
		}
	}
	return votes, nil
}
