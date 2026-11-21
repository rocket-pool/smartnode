package rocketpool

import (
	"context"
	"net/url"
	"strconv"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

func (c *Client) PerformanceWatcherStatus(ctx context.Context) (api.PerformanceWatcherStatusResponse, error) {
	return c.callAPICtx[api.PerformanceWatcherStatusResponse](ctx, "GET", "/api/node/performance-watcher/status", nil, "Could not read performance watcher status")
}
func (c *Client) PerformanceWatcherCandidates(ctx context.Context, offset, limit int) (api.PerformanceWatcherCandidatesResponse, error) {
	return c.callAPICtx[api.PerformanceWatcherCandidatesResponse](ctx, "GET", "/api/node/performance-watcher/candidates", url.Values{"offset": {strconv.Itoa(offset)}, "limit": {strconv.Itoa(limit)}}, "Could not list performance candidates")
}
func (c *Client) PerformanceWatcherCandidate(ctx context.Context, index uint64) (api.PerformanceWatcherCandidateResponse, error) {
	return c.callAPICtx[api.PerformanceWatcherCandidateResponse](ctx, "GET", "/api/node/performance-watcher/candidate", url.Values{"validatorIndex": {strconv.FormatUint(index, 10)}}, "Could not export performance candidate")
}
func (c *Client) PerformanceWatcherValidator(ctx context.Context, index uint64) (api.PerformanceWatcherValidatorResponse, error) {
	return c.callAPICtx[api.PerformanceWatcherValidatorResponse](ctx, "GET", "/api/node/performance-watcher/validator", url.Values{"validatorIndex": {strconv.FormatUint(index, 10)}}, "Could not read collected validator performance")
}

func (c *Client) PerformanceWatcherNodeValidators(ctx context.Context) (api.PerformanceWatcherNodeValidatorsResponse, error) {
	return c.callAPICtx[api.PerformanceWatcherNodeValidatorsResponse](ctx, "GET", "/api/node/performance-watcher/validators", nil, "Could not read collected performance for this node's validators")
}
