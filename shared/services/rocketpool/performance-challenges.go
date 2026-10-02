package rocketpool

import (
	"context"
	"math/big"
	"net/url"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

func (c *Client) GetPerformanceChallenges(id *big.Int) (api.PerformanceChallengesResponse, error) {
	values := url.Values{}
	if id != nil {
		values.Set("challengeId", id.String())
	}
	return c.callAPICtx[api.PerformanceChallengesResponse](context.Background(), "GET", "/api/node/performance-challenges", values, "Could not inspect performance challenges")
}

func (c *Client) CanDefendPerformanceChallenge(id *big.Int) (api.CanActOnPerformanceChallengeResponse, error) {
	return c.callAPICtx[api.CanActOnPerformanceChallengeResponse](context.Background(), "GET", "/api/node/can-defend-performance-challenge", url.Values{"challengeId": {id.String()}}, "Could not check performance challenge defense")
}

func (c *Client) DefendPerformanceChallenge(id *big.Int) (api.PerformanceChallengeResponse, error) {
	return c.callAPICtx[api.PerformanceChallengeResponse](context.Background(), "POST", "/api/node/defend-performance-challenge", url.Values{"challengeId": {id.String()}}, "Could not defend performance challenge")
}

func (c *Client) CanFinalisePerformanceChallenge(id *big.Int) (api.CanActOnPerformanceChallengeResponse, error) {
	return c.callAPICtx[api.CanActOnPerformanceChallengeResponse](context.Background(), "GET", "/api/node/can-finalise-performance-challenge", url.Values{"challengeId": {id.String()}}, "Could not check performance challenge finalisation")
}

func (c *Client) FinalisePerformanceChallenge(id *big.Int) (api.PerformanceChallengeResponse, error) {
	return c.callAPICtx[api.PerformanceChallengeResponse](context.Background(), "POST", "/api/node/finalise-performance-challenge", url.Values{"challengeId": {id.String()}}, "Could not finalise performance challenge")
}
