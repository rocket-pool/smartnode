package rocketpool

import (
	"context"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

type apiResult interface {
	APIError() string
}

// callAPI unmarshals a JSON API reply into T and turns a non-empty Error field into a Go error
func (c *Client) callAPI[T apiResult](method, path string, params url.Values, errPrefix string) (T, error) {
	body, err := c.callHTTPAPI(method, path, params)
	return decodeAPI[T](body, err, errPrefix)
}

// callAPICtx is callAPI with an explicit context (custom timeouts, or none).
func (c *Client) callAPICtx[T apiResult](ctx context.Context, method, path string, params url.Values, errPrefix string) (T, error) {
	body, err := c.callHTTPAPICtx(ctx, method, path, params)
	return decodeAPI[T](body, err, errPrefix)
}

func decodeAPI[T apiResult](body []byte, err error, errPrefix string) (T, error) {
	var zero T
	if err != nil {
		return zero, fmt.Errorf("%s: %w", errPrefix, err)
	}
	var response T
	if err := json.Unmarshal(body, &response); err != nil {
		return zero, fmt.Errorf("%s: could not decode response: %w", errPrefix, err)
	}
	if apiErr := response.APIError(); apiErr != "" {
		return response, fmt.Errorf("%s: %s", errPrefix, apiErr)
	}
	return response, nil
}

// Wait for a transaction — no timeout; blocks until the tx is included or the caller cancels.
func (c *Client) WaitForTransaction(txHash common.Hash) (api.APIResponse, error) {
	return c.callAPICtx[api.APIResponse](context.Background(), "GET", "/api/wait", url.Values{"txHash": {txHash.Hex()}}, "Error waiting for tx")
}
