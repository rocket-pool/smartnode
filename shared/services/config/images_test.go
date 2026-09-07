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
RP_IMAGE_GETH_PROD=ethereum/client-go:v1.17.5@sha256:abc

RP_IMAGE_CURL="curlimages/curl:8.13.0"
RP_IMAGE_ALPINE='alpine:3.21.3'
`)
	got, err := ParseEnvFile(data)
	if err != nil {
		t.Fatal(err)
	}
	if got["RP_IMAGE_GETH_PROD"] != "ethereum/client-go:v1.17.5@sha256:abc" {
		t.Fatalf("geth: %q", got["RP_IMAGE_GETH_PROD"])
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

func TestLoadImagesEmbedded(t *testing.T) {
	catalog, err := LoadImages("")
	if err != nil {
		t.Fatal(err)
	}
	if catalog.Must(ImageGethProd) == "" {
		t.Fatal("empty geth prod")
	}
	if !strings.Contains(catalog.Must(ImageGethProd), "ethereum/client-go:") {
		t.Fatalf("unexpected geth image %s", catalog.Must(ImageGethProd))
	}
}

func TestLoadImagesPrefersDisk(t *testing.T) {
	dir := t.TempDir()
	embedded, err := LoadImages("")
	if err != nil {
		t.Fatal(err)
	}
	values := embedded.Map()
	values[ImageGethProd] = "example.com/geth:custom"
	if err := WriteEnvFile(filepath.Join(dir, ImagesEnvFile), values); err != nil {
		t.Fatal(err)
	}
	catalog, err := LoadImages(dir)
	if err != nil {
		t.Fatal(err)
	}
	if catalog.Must(ImageGethProd) != "example.com/geth:custom" {
		t.Fatalf("got %s", catalog.Must(ImageGethProd))
	}
}

func TestUpdateEnvFilePreservesComments(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ImagesEnvFile)
	if err := os.WriteFile(path, []byte("# comment\nRP_IMAGE_GETH_PROD=old/geth:1\nRP_IMAGE_CURL=keep/curl:1\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := UpdateEnvFile(path, map[string]string{ImageGethProd: "my/geth:custom"}); err != nil {
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
	if parsed[ImageGethProd] != "my/geth:custom" {
		t.Fatalf("geth %q", parsed[ImageGethProd])
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
	if ref != EnvRef(ImageGethProd) {
		t.Fatalf("ec ref %s", ref)
	}
	bn, err := cfg.GetBeaconImageEnvRef()
	if err != nil {
		t.Fatal(err)
	}
	if bn != EnvRef(ImageLighthouseProd) {
		t.Fatalf("bn ref %s", bn)
	}

	cfg.ChangeNetwork(config.Network("testnet"))
	ref, err = cfg.GetECImageEnvRef()
	if err != nil {
		t.Fatal(err)
	}
	if ref != EnvRef(ImageGethTest) {
		t.Fatalf("ec test ref %s", ref)
	}
}

func TestComposeTemplatesUseEnvRefs(t *testing.T) {
	cfg := mustNewRocketPoolConfig(t, "", false)
	cfg.ExecutionClientMode.Value = config.Mode_Local
	cfg.ExecutionClient.Value = config.ExecutionClient_Geth
	cfg.ConsensusClientMode.Value = config.Mode_Local
	cfg.ConsensusClient.Value = config.ConsensusClient_Lighthouse
	cfg.Smartnode.Network.Value = config.Network("mainnet")

	ec, err := cfg.GetECImageEnvRef()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(ec, "${RP_IMAGE_") || !strings.HasSuffix(ec, "}") {
		t.Fatalf("ec ref should be compose interpolation, got %s", ec)
	}
	bn, err := cfg.GetBeaconImageEnvRef()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(bn, "${RP_IMAGE_") {
		t.Fatalf("bn ref %s", bn)
	}
}

func TestSaveImagesEnvWritesTuiOverrides(t *testing.T) {
	dir := t.TempDir()
	embedded, err := LoadImages("")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, ImagesEnvFile)
	if err := WriteEnvFile(path, embedded.Map()); err != nil {
		t.Fatal(err)
	}
	cfg := mustNewRocketPoolConfig(t, dir, false)
	cfg.Smartnode.Network.Value = config.Network("mainnet")
	cfg.Geth.ContainerTag.Value = "my/geth:custom"
	if err := cfg.SaveImagesEnv(); err != nil {
		t.Fatal(err)
	}
	parsed, err := LoadEnvFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if parsed[ImageGethProd] != "my/geth:custom" {
		t.Fatalf("images.env geth %q", parsed[ImageGethProd])
	}
	if cfg.ResolvedImage(ImageGethProd) != "my/geth:custom" {
		t.Fatalf("catalog geth %q", cfg.ResolvedImage(ImageGethProd))
	}

	// Reloading picks the TUI value from images.env, not user-settings.yml.
	reloaded := mustNewRocketPoolConfig(t, dir, false)
	reloaded.Smartnode.Network.Value = config.Network("mainnet")
	if err := reloaded.Geth.ContainerTag.SetToDefault(config.Network("mainnet")); err != nil {
		t.Fatal(err)
	}
	if reloaded.Geth.ContainerTag.Value != "my/geth:custom" {
		t.Fatalf("reloaded geth default %v", reloaded.Geth.ContainerTag.Value)
	}
}

func TestContainerTagsAreNotSerialized(t *testing.T) {
	cfg := mustNewRocketPoolConfig(t, "", false)
	cfg.Geth.ContainerTag.Value = "my/geth:custom"
	serialized := cfg.Serialize()
	if tag, ok := serialized["geth"]["containerTag"]; ok {
		t.Fatalf("container tags must not be saved to user-settings.yml, got %q", tag)
	}
}
