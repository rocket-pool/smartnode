package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rocket-pool/smartnode/shared/types/config"
)

func TestParseEnvFile(t *testing.T) {
	data := []byte(`
# comment
RP_IMAGE_GETH=ethereum/client-go:v1.17.5@sha256:abc

RP_IMAGE_CURL="curlimages/curl:8.13.0"
RP_IMAGE_ALPINE='alpine:3.21.3'
`)
	got, err := ParseEnvFile(data)
	if err != nil {
		t.Fatal(err)
	}
	if got["RP_IMAGE_GETH"] != "ethereum/client-go:v1.17.5@sha256:abc" {
		t.Fatalf("geth: %q", got["RP_IMAGE_GETH"])
	}
	if got["RP_IMAGE_CURL"] != "curlimages/curl:8.13.0" {
		t.Fatalf("curl: %q", got["RP_IMAGE_CURL"])
	}
	if got["RP_IMAGE_ALPINE"] != "alpine:3.21.3" {
		t.Fatalf("alpine: %q", got["RP_IMAGE_ALPINE"])
	}
}

func TestParseEnvFileRejectsBadLine(t *testing.T) {
	if _, err := ParseEnvFile([]byte("not-a-key-value")); err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadImageCatalogsEmbedded(t *testing.T) {
	mainnet, overlays, err := LoadImageCatalogs("", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(mainnet.Must(ImageGeth), "ethereum/client-go:") {
		t.Fatalf("unexpected geth image %s", mainnet.Must(ImageGeth))
	}
	if len(overlays) != 0 {
		t.Fatalf("expected no overlays without networks, got %v", overlays)
	}
}

func TestLoadImageCatalogsPrefersDisk(t *testing.T) {
	dir := t.TempDir()
	mainnet, _, err := LoadImageCatalogs("", nil)
	if err != nil {
		t.Fatal(err)
	}
	values := mainnet.Map()
	values[ImageGeth] = "example.com/geth:custom"
	if err := WriteEnvFile(filepath.Join(dir, ImagesMainnetFile), values); err != nil {
		t.Fatal(err)
	}
	loaded, _, err := LoadImageCatalogs(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Must(ImageGeth) != "example.com/geth:custom" {
		t.Fatalf("got %s", loaded.Must(ImageGeth))
	}
}

func TestNetworkOverlayOverridesMainnet(t *testing.T) {
	dir := t.TempDir()
	mainnet, _, err := LoadImageCatalogs("", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteEnvFile(filepath.Join(dir, ImagesMainnetFile), mainnet.Map()); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ImagesTestnetFile), []byte("RP_IMAGE_GETH=example.com/geth:testnet\n"), 0644); err != nil {
		t.Fatal(err)
	}
	networks, err := LoadNetworks("")
	if err != nil {
		t.Fatal(err)
	}
	cfg := mustNewRocketPoolConfig(t, dir, false)
	cfg.networks = networks
	cfg.Smartnode.Network.Value = config.Network("testnet")
	if got := cfg.imageForNetwork(ImageGeth, config.Network("testnet")); got != "example.com/geth:testnet" {
		t.Fatalf("testnet geth %s", got)
	}
	if got := cfg.imageForNetwork(ImageLighthouse, config.Network("testnet")); got != mainnet.Must(ImageLighthouse) {
		t.Fatalf("testnet lighthouse should inherit mainnet, got %s", got)
	}
	if got := cfg.imageForNetwork(ImageGeth, config.Network("mainnet")); got != mainnet.Must(ImageGeth) {
		t.Fatalf("mainnet geth should be unchanged, got %s", got)
	}
}

func TestUpdateEnvFilePreservesComments(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ImagesMainnetFile)
	if err := os.WriteFile(path, []byte("# comment\nRP_IMAGE_GETH=old/geth:1\nRP_IMAGE_CURL=keep/curl:1\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := UpdateEnvFile(path, map[string]string{ImageGeth: "my/geth:custom"}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.Contains(text, "# comment") {
		t.Fatalf("lost comment: %s", text)
	}
	parsed, err := ParseEnvFile(data)
	if err != nil {
		t.Fatal(err)
	}
	if parsed[ImageGeth] != "my/geth:custom" {
		t.Fatalf("geth %q", parsed[ImageGeth])
	}
	if parsed[ImageCurl] != "keep/curl:1" {
		t.Fatalf("curl %q", parsed[ImageCurl])
	}
}

func TestClientImageEnvRefs(t *testing.T) {
	cfg := mustNewRocketPoolConfig(t, "", false)
	cfg.ExecutionClientMode.Value = config.Mode_Local
	cfg.ExecutionClient.Value = config.ExecutionClient_Geth
	cfg.ConsensusClientMode.Value = config.Mode_Local
	cfg.ConsensusClient.Value = config.ConsensusClient_Lighthouse
	cfg.Smartnode.Network.Value = config.Network("mainnet")

	ref, err := cfg.GetECImageEnvRef()
	if err != nil {
		t.Fatal(err)
	}
	if ref != ImageTagRef(ECImageTagOverride, ECImageTagDefault) {
		t.Fatalf("ec ref %s", ref)
	}
	bn, err := cfg.GetBeaconImageEnvRef()
	if err != nil {
		t.Fatal(err)
	}
	if bn != ImageTagRef(BNImageTagOverride, BNImageTagDefault) {
		t.Fatalf("bn ref %s", bn)
	}
	vc, err := cfg.GetVCImageEnvRef()
	if err != nil {
		t.Fatal(err)
	}
	if vc != ImageTagRef(VCImageTagOverride, VCImageTagDefault) {
		t.Fatalf("vc ref %s", vc)
	}

	cfg.ChangeNetwork(config.Network("testnet"))
	ref, err = cfg.GetECImageEnvRef()
	if err != nil {
		t.Fatal(err)
	}
	if ref != ImageTagRef(ECImageTagOverride, ECImageTagDefault) {
		t.Fatalf("ec test ref %s", ref)
	}
	files := cfg.ImagesEnvFiles()
	if len(files) < 1 || files[0] != ImagesMainnetFile {
		t.Fatalf("expected mainnet.env first, got %v", files)
	}
}

func TestComposeEnvOverrides(t *testing.T) {
	cfg := mustNewRocketPoolConfig(t, "", false)
	cfg.Smartnode.Network.Value = config.Network("mainnet")
	if err := cfg.Geth.ContainerTag.SetToDefault(config.Network("mainnet")); err != nil {
		t.Fatal(err)
	}
	if len(cfg.ComposeEnvOverrides()) != 0 {
		t.Fatalf("expected no overrides, got %v", cfg.ComposeEnvOverrides())
	}
	cfg.Geth.ContainerTag.Value = "my/geth:custom"
	overrides := cfg.ComposeEnvOverrides()
	if overrides[ECImageTagOverride] != "my/geth:custom" {
		t.Fatalf("got %v", overrides)
	}
	defaults := cfg.ComposeImageDefaults()
	if defaults[ECImageTagDefault] == "" || defaults[ECImageTagDefault] == "my/geth:custom" {
		t.Fatalf("default should stay the catalog pin, got %q", defaults[ECImageTagDefault])
	}
	env, err := cfg.ComposeImageEnv()
	if err != nil {
		t.Fatal(err)
	}
	if env[ECImageTagDefault] != defaults[ECImageTagDefault] {
		t.Fatalf("merged default %q", env[ECImageTagDefault])
	}
	if env[ECImageTagOverride] != "my/geth:custom" {
		t.Fatalf("merged override %q", env[ECImageTagOverride])
	}
}

func TestComposeEnvAssignmentsSorted(t *testing.T) {
	cfg := mustNewRocketPoolConfig(t, "", false)
	cfg.Smartnode.Network.Value = config.Network("mainnet")
	cfg.ExecutionClientMode.Value = config.Mode_Local
	cfg.ExecutionClient.Value = config.ExecutionClient_Geth
	cfg.Geth.ContainerTag.Value = "my/geth:custom"
	cfg.Prometheus.ContainerTag.Value = "my/prom:custom"
	got, err := cfg.ComposeEnvAssignments()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) < 2 {
		t.Fatalf("expected multiple assignments, got %v", got)
	}
	for i := 1; i < len(got); i++ {
		if got[i-1] >= got[i] {
			t.Fatalf("assignments not sorted: %v", got)
		}
	}
}

func TestComposeEnvAssignmentsRejectsWhitespace(t *testing.T) {
	cfg := mustNewRocketPoolConfig(t, "", false)
	cfg.Smartnode.Network.Value = config.Network("mainnet")
	cfg.ExecutionClientMode.Value = config.Mode_Local
	cfg.ExecutionClient.Value = config.ExecutionClient_Geth
	cfg.Geth.ContainerTag.Value = "my/geth custom"
	if _, err := cfg.ComposeEnvAssignments(); err == nil {
		t.Fatal("expected error for whitespace in override value")
	}
}

func TestContainerTagsAreSerialized(t *testing.T) {
	cfg := mustNewRocketPoolConfig(t, "", false)
	cfg.Geth.ContainerTag.Value = "my/geth:custom"
	serialized := cfg.Serialize()
	if tag, ok := serialized["geth"]["containerTag"]; !ok || tag != "my/geth:custom" {
		t.Fatalf("container tags should be saved to user-settings.yml, got %v", serialized["geth"]["containerTag"])
	}
}
