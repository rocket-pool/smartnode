package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// WatcherRequest is a cancellable, bounded read-only transport. The watcher
// validates response metadata and handles historical-data fallback itself.
func (c *StandardHttpClient) WatcherRequest(ctx context.Context, path string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.providerAddress+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("beacon %s: HTTP %d", path, resp.StatusCode)
	}
	const limit = 128 << 20 // Includes large committee responses, never full states.
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err == nil && len(data) > limit {
		err = fmt.Errorf("beacon response exceeds watcher limit")
	}
	return data, err
}
