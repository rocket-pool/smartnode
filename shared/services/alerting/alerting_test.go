package alerting

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// makeTestConfig returns a RocketPoolConfig pointed at a local httptest server.
// It enables alerting and sets the native-mode host/port from the server URL.
func makeTestConfig(serverURL string) (*config.RocketPoolConfig, error) {
	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, fmt.Errorf("parsing test server URL: %w", err)
	}
	port, err := strconv.ParseUint(u.Port(), 10, 16)
	if err != nil {
		return nil, fmt.Errorf("parsing test server port: %w", err)
	}

	cfg := config.NewRocketPoolConfig("", true /* isNativeMode */)
	cfg.Alertmanager.EnableAlerting.Value = true
	cfg.Alertmanager.NativeModeHost.Value = u.Hostname()
	cfg.Alertmanager.NativeModePort.Value = uint16(port)
	return cfg, nil
}

// lowETHAlertJSON returns a minimal Alertmanager /api/v2/alerts JSON response
// containing one active LowETHBalance alert.
func lowETHAlertJSON() string {
	return `[{
		"labels":      {"alertname": "LowETHBalance", "severity": "critical", "job": "node"},
		"annotations": {"summary": "Low ETH Balance", "description": "The node ETH balance is low."},
		"status":      {"state": "active", "silencedBy": [], "inhibitedBy": []},
		"receivers":   [{"name": "node_operator_default"}],
		"fingerprint": "abc123",
		"startsAt":    "2026-01-01T00:00:00Z",
		"endsAt":      "2099-01-01T00:00:00Z",
		"updatedAt":   "2026-01-01T00:00:00Z"
	}]`
}

func TestFetchAlerts_LowETHBalance(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/alerts" || r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, lowETHAlertJSON())
	}))
	defer srv.Close()

	cfg, err := makeTestConfig(srv.URL)
	if err != nil {
		t.Fatalf("failed to build config: %v", err)
	}

	alerts, err := FetchAlerts(cfg)
	if err != nil {
		t.Fatalf("FetchAlerts returned error: %v", err)
	}
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}

	a := alerts[0]
	if got := a.Labels["alertname"]; got != "LowETHBalance" {
		t.Errorf("alertname: want LowETHBalance, got %q", got)
	}
	if got := a.Labels["severity"]; got != "critical" {
		t.Errorf("severity: want critical, got %q", got)
	}
	if got := *a.Status.State; got != "active" {
		t.Errorf("state: want active, got %q", got)
	}
}

func TestFetchAlerts_AlertingDisabled(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		fmt.Fprint(w, lowETHAlertJSON())
	}))
	defer srv.Close()

	cfg, err := makeTestConfig(srv.URL)
	if err != nil {
		t.Fatalf("failed to build config: %v", err)
	}
	cfg.Alertmanager.EnableAlerting.Value = false

	alerts, err := FetchAlerts(cfg)
	if err != nil {
		t.Fatalf("FetchAlerts returned unexpected error: %v", err)
	}
	if alerts != nil {
		t.Errorf("expected nil alerts when alerting disabled, got %v", alerts)
	}
	if called {
		t.Error("HTTP server should not have been contacted when alerting is disabled")
	}
}

func TestFetchAlerts_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	cfg, err := makeTestConfig(srv.URL)
	if err != nil {
		t.Fatalf("failed to build config: %v", err)
	}

	_, err = FetchAlerts(cfg)
	if err == nil {
		t.Error("expected an error when the server returns 500, got nil")
	}
}

// TestNodeAlert_LowETHBalance tests the NodeAlert helper methods using a
// LowETHBalance alert as representative input.
func TestNodeAlert_LowETHBalance(t *testing.T) {
	alert := api.NodeAlert{
		State: "active",
		Labels: map[string]string{
			"alertname": "LowETHBalance",
			"severity":  "critical",
		},
		Annotations: map[string]string{
			"summary":     "Low ETH Balance",
			"description": "The node ETH balance is low.",
		},
	}

	if !alert.IsActive() {
		t.Error("expected IsActive() == true")
	}
	if alert.IsSuppressed() {
		t.Error("expected IsSuppressed() == false")
	}
	if got := alert.Severity(); got != "critical" {
		t.Errorf("Severity(): want critical, got %q", got)
	}
	if got := alert.Summary(); got != "Low ETH Balance" {
		t.Errorf("Summary(): want %q, got %q", "Low ETH Balance", got)
	}
	if got := alert.Description(); got != "The node ETH balance is low." {
		t.Errorf("Description(): want %q, got %q", "The node ETH balance is low.", got)
	}

	colored := alert.ColorString()
	if colored == "" {
		t.Error("ColorString() returned empty string")
	}
	// Critical alerts use the red ANSI code.
	const red = "\033[31m"
	if len(colored) < len(red) || colored[:len(red)] != red {
		t.Errorf("ColorString() for critical alert should start with red ANSI code, got %q", colored)
	}
}
