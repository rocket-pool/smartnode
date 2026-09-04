package services

import (
	"errors"
	"testing"

	"github.com/fatih/color"

	log "github.com/rocket-pool/smartnode/shared/logger"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/beacon/client"
)

func testLogger() log.ColorLogger {
	return log.NewColorLogger(color.FgHiBlue)
}

func TestBeaconRunFunction1PrefersFallback(t *testing.T) {
	t.Parallel()

	primary := client.NewStandardHttpClient("http://primary")
	fallback := client.NewStandardHttpClient("http://fallback")
	m := &BeaconClientManager{
		primaryBc:      primary,
		fallbackBc:     fallback,
		primaryReady:   true,
		fallbackReady:  true,
		preferFallback: true,
	}

	var used beacon.Client
	result, err := m.runFunction1(func(c beacon.Client) (interface{}, error) {
		used = c
		return "ok", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if result != "ok" {
		t.Fatalf("result = %v, want ok", result)
	}
	if used != fallback {
		t.Fatal("expected fallback client to be used first")
	}
}

func TestBeaconRunFunction1FallbackDisconnectUsesPrimary(t *testing.T) {
	t.Parallel()

	primary := client.NewStandardHttpClient("http://primary")
	fallback := client.NewStandardHttpClient("http://fallback")
	m := &BeaconClientManager{
		primaryBc:      primary,
		fallbackBc:     fallback,
		primaryReady:   true,
		fallbackReady:  true,
		preferFallback: true,
		logger:         testLogger(),
	}

	result, err := m.runFunction1(func(c beacon.Client) (interface{}, error) {
		if c == fallback {
			return nil, errors.New("dial tcp fallback: connection refused")
		}
		return "primary", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if result != "primary" {
		t.Fatalf("result = %v, want primary", result)
	}
	if m.fallbackReady {
		t.Fatal("fallback should be marked not ready after disconnect")
	}
	if !m.primaryReady {
		t.Fatal("primary should still be ready")
	}
}

func TestBeaconRunFunction1StaticShortCircuits(t *testing.T) {
	t.Parallel()

	static := client.NewStandardHttpClient("http://static")
	fallback := client.NewStandardHttpClient("http://fallback")
	m := NewStaticBeaconClientManager(static)
	m.preferFallback = true
	m.fallbackReady = true
	m.fallbackBc = fallback

	var used beacon.Client
	_, err := m.runFunction1(func(c beacon.Client) (interface{}, error) {
		used = c
		return nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if used != static {
		t.Fatal("static manager should not route through primary/fallback")
	}
}

func TestExecutionRunFunctionPrefersFallback(t *testing.T) {
	t.Parallel()

	primary := &EthClient{}
	fallback := &EthClient{}
	p := &ExecutionClientManager{
		primaryEc:      primary,
		fallbackEc:     fallback,
		primaryReady:   true,
		fallbackReady:  true,
		preferFallback: true,
	}

	var used *EthClient
	result, err := p.runFunction(func(c *EthClient) (interface{}, error) {
		used = c
		return "ok", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if result != "ok" {
		t.Fatalf("result = %v, want ok", result)
	}
	if used != fallback {
		t.Fatal("expected fallback client to be used first")
	}
}
