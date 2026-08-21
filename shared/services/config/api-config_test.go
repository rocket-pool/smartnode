package config

import (
	"path/filepath"
	"testing"

	"github.com/rocket-pool/smartnode/shared/services/config/migration"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

func TestGetNodeOpenPorts(t *testing.T) {
	cfg := NewRocketPoolConfig("/tmp/rp-test", false)

	cfg.Api.OpenApiPort.Value = cfgtypes.RPC_Closed
	if got := cfg.GetNodeOpenPorts(); got != "" {
		t.Fatalf("closed: %q", got)
	}

	cfg.Api.OpenApiPort.Value = cfgtypes.RPC_OpenLocalhost
	cfg.Api.ApiPort.Value = uint16(8280)
	got := cfg.GetNodeOpenPorts()
	if got != `"127.0.0.1:8280:8280/tcp"` {
		t.Fatalf("localhost: %q", got)
	}

	cfg.Api.OpenApiPort.Value = cfgtypes.RPC_OpenExternal
	got = cfg.GetNodeOpenPorts()
	if got != `"8280:8280/tcp"` {
		t.Fatalf("external: %q", got)
	}
}

func TestTokenPathExpandsTilde(t *testing.T) {
	got := tokenPath("~/.rocketpool/data")
	if got[0] == '~' {
		t.Fatalf("tilde was not expanded: %q", got)
	}
	if filepath.Base(got) != "api-token" {
		t.Fatalf("unexpected token file %q", got)
	}
}

func TestDefaultRateLimit(t *testing.T) {
	cfg := NewRocketPoolConfig("/tmp/rp-test", false)
	got, ok := cfg.Api.RateLimit.Value.(uint16)
	if !ok || got != 5 {
		t.Fatalf("default rate limit %v (%T), want 5", cfg.Api.RateLimit.Value, cfg.Api.RateLimit.Value)
	}
}

func TestSensitiveTokenNotSerialized(t *testing.T) {
	cfg := NewRocketPoolConfig("/tmp/rp-test", false)
	cfg.Api.APIToken.Value = "rpsn_secret"
	serialized := cfg.Serialize()
	if _, exists := serialized["api"]["apiToken"]; exists {
		t.Fatal("API token must not be written to user-settings.yml")
	}
}

func TestMigrateApiPort(t *testing.T) {
	serialized := map[string]map[string]string{
		"root": {
			"version":  "v1.21.0",
			"isNative": "false",
			"rpDir":    "/tmp",
		},
		"smartnode": {
			"apiPort": "9001",
			"network": "mainnet",
		},
	}
	if err := migration.UpdateConfig(serialized); err != nil {
		t.Fatal(err)
	}
	if serialized["smartnode"]["apiPort"] != "" {
		t.Fatalf("old key still present: %q", serialized["smartnode"]["apiPort"])
	}
	if serialized["api"]["apiPort"] != "9001" {
		t.Fatalf("migrated port %q", serialized["api"]["apiPort"])
	}
}
