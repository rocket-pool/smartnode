package client

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type watcherRoundTripper func(*http.Request) (*http.Response, error)

func (f watcherRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestWatcherTransportContextAndEndpoint(t *testing.T) {
	previous := http.DefaultClient
	defer func() { http.DefaultClient = previous }()
	http.DefaultClient = &http.Client{Transport: watcherRoundTripper(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/eth/v2/beacon/blocks/0x123/attestations" || req.Method != http.MethodGet || req.Header.Get("Accept") != "application/json" {
			t.Fatal(req)
		}
		if _, ok := req.Context().Deadline(); !ok {
			t.Fatal("missing timeout")
		}
		if err := req.Context().Err(); err != nil {
			return nil, err
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"data":[]}`)), Header: make(http.Header)}, nil
	})}
	client := NewStandardHttpClient("http://beacon")
	data, err := client.WatcherRequest(context.Background(), "/eth/v2/beacon/blocks/0x123/attestations")
	if err != nil || string(data) != `{"data":[]}` {
		t.Fatal(string(data), err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = client.WatcherRequest(ctx, "/eth/v2/beacon/blocks/0x123/attestations"); err == nil {
		t.Fatal("canceled request succeeded")
	}
}
