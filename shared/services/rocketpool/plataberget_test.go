package rocketpool

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/rocket-pool/smartnode/shared/services/config"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
	"gopkg.in/yaml.v2"
)

// Rendering must not query public IP services during a unit test.
type offlineNetworkTemplateConfig struct{ *config.RocketPoolConfig }

func (offlineNetworkTemplateConfig) GetExternalIp() string { return "192.0.2.1" }

func TestPlatabergetNamedNetworkTemplates(t *testing.T) {
	cfg, err := config.NewRocketPoolConfig(t.TempDir(), false)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ChangeNetwork("plataberget")
	// Match service startup, which loads the serialized choice types from disk.
	if err := cfg.Deserialize(cfg.Serialize()); err != nil {
		t.Fatal(err)
	}
	for _, cc := range []cfgtypes.ConsensusClient{cfgtypes.ConsensusClient_Lodestar, cfgtypes.ConsensusClient_Nimbus, cfgtypes.ConsensusClient_Teku} {
		t.Run(string(cc), func(t *testing.T) {
			cfg.ConsensusClient.Value = cc
			for _, service := range []string{"eth1", "eth2", "validator"} {
				body, err := os.ReadFile(filepath.Join("assets", "install", "templates", service+".tmpl"))
				if err != nil {
					t.Fatal(err)
				}
				tmpl, err := template.New(service).Parse(string(body))
				if err != nil {
					t.Fatal(err)
				}
				var rendered bytes.Buffer
				if err := tmpl.Execute(&rendered, offlineNetworkTemplateConfig{cfg}); err != nil {
					t.Fatal(err)
				}
				var compose map[string]any
				if err := yaml.Unmarshal(rendered.Bytes(), &compose); err != nil {
					t.Fatalf("invalid %s YAML: %v", service, err)
				}
				text := rendered.String()
				if !strings.Contains(text, "BEACON_NETWORK=plataberget") || strings.Contains(text, "CUSTOM_CHAIN_DIR") || strings.Contains(text, "/custom-chain") {
					t.Fatalf("%s did not use the named preset", service)
				}
				client := "CLIENT=nethermind"
				if service != "eth1" {
					client = "CC_CLIENT=" + string(cc)
				}
				if !strings.Contains(text, client) || strings.Contains(text, "devnet") {
					t.Fatalf("unexpected client or devnet startup in %s", service)
				}
			}
		})
	}
}
