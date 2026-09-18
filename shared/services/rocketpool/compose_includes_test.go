package rocketpool

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/alessio/shellescape"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool/assets"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
	"gopkg.in/yaml.v2"
)

func writeComposeTestFile(t *testing.T, project, name, content string) string {
	t.Helper()
	path := filepath.Join(project, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func newComposeTestProject(t *testing.T, network string, overlay bool) (*config.RocketPoolConfig, string) {
	t.Helper()
	project := filepath.Join(t.TempDir(), "rocket pool")
	src, err := os.ReadFile(filepath.Join("assets", "install", "templates", "compose.tmpl"))
	if err != nil {
		t.Fatal(err)
	}
	writeComposeTestFile(t, project, "templates/compose.tmpl", string(src))
	writeComposeTestFile(t, project, "mainnet.env", string(assets.ImagesMainnetEnv()))
	writeComposeTestFile(t, project, "override/compose.yml", "{}\n")
	if err := os.MkdirAll(filepath.Join(project, runtimeDir), 0755); err != nil {
		t.Fatal(err)
	}
	if network == "custom" {
		writeComposeTestFile(t, project, "networks-extra.yml", `version: 1
networks:
  - name: custom
    label: Custom
    description: Test network
    chainID: 12345
    beaconNetwork: custom
    clientTagSet: test
`)
	}
	if overlay {
		writeComposeTestFile(t, project, network+".env",
			"RP_IMAGE_GETH="+testImage("geth-network")+"\nRP_IMAGE_NIMBUS_VC="+testImage("vc-network")+"\n")
	}
	cfg, err := config.NewRocketPoolConfig(project, false)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Smartnode.Network.Value = cfgtypes.Network(network)
	cfg.ExecutionClientMode.Value = cfgtypes.Mode_Local
	cfg.ExecutionClient.Value = cfgtypes.ExecutionClient_Geth
	cfg.ConsensusClientMode.Value = cfgtypes.Mode_Local
	cfg.ConsensusClient.Value = cfgtypes.ConsensusClient_Nimbus
	// Select the network defaults, as the TUI does when switching networks.
	if err := cfg.Geth.ContainerTag.SetToDefault(cfg.GetNetwork()); err != nil {
		t.Fatal(err)
	}
	if err := cfg.Nimbus.BnContainerTag.SetToDefault(cfg.GetNetwork()); err != nil {
		t.Fatal(err)
	}
	if err := cfg.Nimbus.VcContainerTag.SetToDefault(cfg.GetNetwork()); err != nil {
		t.Fatal(err)
	}
	return cfg, project
}

func testImage(name string) string {
	return "example.com/" + name + ":test@sha256:" + strings.Repeat("a", 64)
}

type testComposeInclude struct {
	Paths      []string `yaml:"path"`
	ProjectDir string   `yaml:"project_directory"`
	EnvFiles   []string `yaml:"env_file"`
}

func readComposeInclude(t *testing.T, path string) testComposeInclude {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var model struct {
		Includes []testComposeInclude `yaml:"include"`
	}
	if err := yaml.Unmarshal(b, &model); err != nil {
		t.Fatal(err)
	}
	if len(model.Includes) != 1 {
		t.Fatalf("need one merge scope, got %d includes", len(model.Includes))
	}
	return model.Includes[0]
}

func requireDockerCompose(t *testing.T) {
	t.Helper()
	if out, err := exec.Command("docker", "compose", "version").CombinedOutput(); err != nil {
		t.Skipf("Docker Compose unavailable: %v\n%s", err, out)
	}
}

func renderComposeTest(t *testing.T, cfg *config.RocketPoolConfig, args ...string) map[string]interface{} {
	t.Helper()
	assignments, err := cfg.ComposeEnvAssignments()
	if err != nil {
		t.Fatal(err)
	}
	command := append([]string{"docker", "compose", "--project-directory", cfg.RocketPoolDirectory, "--project-name", "rocketpool"}, args...)
	command = append(command, "config", "--format", "json")
	for i, arg := range command {
		command[i] = shellescape.Quote(arg)
	}
	cmd := exec.Command("sh", "-c", strings.Join(append(assignments, strings.Join(command, " ")), " "))
	// Make interpolation independent of the developer's shell and Compose setup.
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if !strings.HasPrefix(key, "COMPOSE_") && !strings.HasPrefix(key, "RP_IMAGE_") && !strings.Contains(key, "_IMAGE_TAG_") {
			cmd.Env = append(cmd.Env, entry)
		}
	}
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			t.Fatalf("compose config: %v\n%s", err, exitErr.Stderr)
		}
		t.Fatal(err)
	}
	var model map[string]interface{}
	if err := json.Unmarshal(out, &model); err != nil {
		t.Fatalf("decode config: %v\n%s", err, out)
	}
	return model
}

// This runs without Docker and verifies file ordering, optional overlays, and
// CLI paths relative to the caller rather than the include's project directory.
func TestComposeFileOrderAndEnvFiles(t *testing.T) {
	for _, network := range []string{"mainnet", "testnet", "devnet", "custom"} {
		t.Run(network, func(t *testing.T) {
			cfg, project := newComposeTestProject(t, network, false)
			deployed := []string{
				filepath.Join(project, "runtime/node.yml"), filepath.Join(project, "override/node.yml"),
				filepath.Join(project, "runtime/eth1.yml"), filepath.Join(project, "override/eth1.yml"),
				filepath.Join(project, "runtime/addons/gww/addon_gww.yml"), filepath.Join(project, "override/addons/gww/addon_gww.yml"),
			}
			path, err := writeComposeFile(cfg, project, deployed, []string{"custom override.yml"})
			if err != nil {
				t.Fatal(err)
			}
			got := readComposeInclude(t, path)
			extra, err := filepath.Abs("custom override.yml")
			if err != nil {
				t.Fatal(err)
			}
			wantPaths := append(append([]string{}, deployed...), filepath.Join(project, "override/compose.yml"), extra)
			wantEnv := []string{filepath.Join(project, "mainnet.env")}
			if network == "testnet" || network == "devnet" {
				wantEnv = append(wantEnv, filepath.Join(project, network+".env"))
			}
			wantEnv = append(wantEnv, filepath.Join(project, "runtime", config.ComposeImageEnvFile))
			want := testComposeInclude{Paths: wantPaths, ProjectDir: project, EnvFiles: wantEnv}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("got %#v\nwant %#v", got, want)
			}
		})
	}
}

func TestComposeImageEnvPrecedence(t *testing.T) {
	requireDockerCompose(t)
	for _, network := range []string{"mainnet", "testnet", "devnet", "custom", "custom-no-overlay"} {
		for _, tui := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/tui=%t", network, tui), func(t *testing.T) {
				selected := strings.TrimSuffix(network, "-no-overlay")
				overlay := network != "mainnet" && network != "custom-no-overlay"
				cfg, project := newComposeTestProject(t, selected, overlay)
				base := writeComposeTestFile(t, project, "runtime/clients.yml", `services:
  eth1:
    image: ${EC_IMAGE_TAG_OVERRIDE:-${EC_IMAGE_TAG_DEFAULT}}
  eth2:
    image: ${BN_IMAGE_TAG_OVERRIDE:-${BN_IMAGE_TAG_DEFAULT}}
  validator:
    image: ${VC_IMAGE_TAG_OVERRIDE:-${VC_IMAGE_TAG_DEFAULT}}
  grafana:
    image: ${GRAFANA_IMAGE_TAG_OVERRIDE:-${GRAFANA_IMAGE_TAG_DEFAULT}}
`)
				wantEC := cfg.ResolvedImage(config.ImageGeth)
				wantVC := cfg.ResolvedImage(config.ImageNimbusVc)
				if tui {
					// The TUI must also beat an explicit *_OVERRIDE in the env files.
					envName := "mainnet.env"
					if overlay {
						envName = selected + ".env"
					}
					content, err := os.ReadFile(filepath.Join(project, envName))
					if err != nil {
						t.Fatal(err)
					}
					writeComposeTestFile(t, project, envName, string(content)+"\nVC_IMAGE_TAG_OVERRIDE="+testImage("vc-env-override")+"\n")
					wantEC, wantVC = testImage("geth-tui"), testImage("vc-tui")
					cfg.Geth.ContainerTag.Value, cfg.Nimbus.VcContainerTag.Value = wantEC, wantVC
				}
				path, err := writeComposeFile(cfg, project, []string{base}, nil)
				if err != nil {
					t.Fatal(err)
				}
				got := renderComposeTest(t, cfg, "-f", path)["services"].(map[string]interface{})
				expected := map[string]string{
					"eth1": wantEC, "validator": wantVC,
					"eth2": cfg.ResolvedImage(config.ImageNimbusBn), "grafana": cfg.ResolvedImage(config.ImageGrafana),
				}
				for service, image := range expected {
					if actual := got[service].(map[string]interface{})["image"]; actual != image {
						t.Errorf("%s image = %v, want %s", service, actual, image)
					}
				}
				if !tui {
					// Prove Compose reads the env files, instead of values frozen in Go's
					// process environment when the configuration was loaded.
					writeComposeTestFile(t, project, "mainnet.env", string(assets.ImagesMainnetEnv())+"\nRP_IMAGE_GRAFANA="+testImage("grafana-updated")+"\n")
					updated := renderComposeTest(t, cfg, "-f", path)["services"].(map[string]interface{})
					if image := updated["grafana"].(map[string]interface{})["image"]; image != testImage("grafana-updated") {
						t.Fatalf("env_file edit did not take effect: %v", image)
					}
				}
			})
		}
	}
}

func TestComposeOverridesMatchOrderedFiles(t *testing.T) {
	requireDockerCompose(t)
	cfg, project := newComposeTestProject(t, "mainnet", false)
	node := writeComposeTestFile(t, project, "runtime/node.yml", `services:
  node:
    image: ${SMARTNODE_IMAGE_TAG_OVERRIDE:-${SMARTNODE_IMAGE_TAG_DEFAULT}}
    labels: {example: original}
    networks: [net]
networks:
  net:
    driver_opts: {com.docker.network.driver.mtu: "1500"}
`)
	nodeOverride := writeComposeTestFile(t, project, "override/node.yml", "{}\n")
	base := writeComposeTestFile(t, project, "runtime/eth1.yml", `services:
  eth1:
    image: ${EC_IMAGE_TAG_OVERRIDE:-${EC_IMAGE_TAG_DEFAULT}}
    user: root
    restart: unless-stopped
    command: ["original"]
    entrypoint: ["original-entry"]
    environment: {KEY: original}
    labels: {example: original}
    healthcheck:
      test: ["CMD", "original"]
      interval: 30s
    ports:
      - target: 8545
        published: "8545"
        name: original
    volumes:
      - eth1clientdata:/ethclient
      - ./scripts:/setup:ro
    secrets: [{source: oldsecret, target: item}]
    configs: [{source: oldconfig, target: /etc/item}]
    networks: [net]
networks:
  net: {}
volumes:
  eth1clientdata: {}
secrets:
  oldsecret: {external: true}
  newsecret: {external: true}
configs:
  oldconfig: {external: true}
  newconfig: {external: true}
`)
	cases := []struct{ name, body string }{
		{"image", "    image: example.com/custom:tag\n"},
		{"user", "    user: '1000:1000'\n"},
		{"restart", "    restart: always\n"},
		{"command", "    command: ['custom']\n"},
		{"entrypoint", "    entrypoint: ['custom-entry']\n"},
		{"environment", "    environment: {KEY: custom, NEW_KEY: added}\n"},
		{"labels", "    labels: {example: custom, added: value}\n"},
		{"healthcheck", "    healthcheck: {test: ['CMD', 'custom'], interval: 60s}\n"},
		{"volumes-short", "    volumes: ['/tmp/custom:/ethclient', './ancient:/ethclient/geth/chaindata/ancient']\n"},
		{"volumes-long", "    volumes:\n      - {type: bind, source: /tmp/custom, target: /ethclient}\n      - {type: bind, source: ./ancient, target: /ethclient/geth/chaindata/ancient}\n"},
		{"ports-same-key", "    ports: [{target: 8545, published: '8545', name: custom}]\n"},
		{"ports-additive", "    ports: ['9545:8545']\n"},
		{"secrets", "    secrets: [{source: newsecret, target: item}]\n"},
		{"configs", "    configs: [{source: newconfig, target: /etc/item}]\n"},
		{"reset-ports", "    ports: !reset []\n"},
		{"replace-ports", "    ports: !override ['9545:8545']\n"},
		{"cross-service-and-network", "    labels: {example: custom}\n  node:\n    labels: {example: custom}\nnetworks:\n  net:\n    driver_opts: {com.docker.network.driver.mtu: '1400'}\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			override := writeComposeTestFile(t, project, "override/eth1.yml", "services:\n  eth1:\n"+tc.body)
			deployed := []string{node, nodeOverride, base, override}
			path, err := writeComposeFile(cfg, project, deployed, nil)
			if err != nil {
				t.Fatal(err)
			}
			assertComposeMatchesFiles(t, cfg, path)
		})
	}
	t.Run("global-and-cli-precedence", func(t *testing.T) {
		override := writeComposeTestFile(t, project, "override/eth1.yml", "services:\n  eth1:\n    environment: {KEY: service, SERVICE_ONLY: kept}\n")
		writeComposeTestFile(t, project, "override/compose.yml", "services:\n  eth1:\n    environment: {KEY: global, GLOBAL_ONLY: kept}\n    ports: !reset []\n")
		extra := writeComposeTestFile(t, t.TempDir(), "cli override.yml", "services:\n  eth1:\n    environment: {KEY: cli}\n    command: ['cli']\n")
		cwd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		relative, err := filepath.Rel(cwd, extra)
		if err != nil {
			t.Fatal(err)
		}
		path, err := writeComposeFile(cfg, project, []string{node, nodeOverride, base, override}, []string{relative})
		if err != nil {
			t.Fatal(err)
		}
		assertComposeMatchesFiles(t, cfg, path)
	})
}

func assertComposeMatchesFiles(t *testing.T, cfg *config.RocketPoolConfig, path string) {
	t.Helper()
	include := readComposeInclude(t, path)
	var args []string
	for _, env := range include.EnvFiles {
		args = append(args, "--env-file", env)
	}
	for _, file := range include.Paths {
		args = append(args, "-f", file)
	}
	got := renderComposeTest(t, cfg, "-f", path)
	want := renderComposeTest(t, cfg, args...)
	if !reflect.DeepEqual(got, want) {
		actual, _ := json.MarshalIndent(got, "", "  ")
		expected, _ := json.MarshalIndent(want, "", "  ")
		t.Fatalf("single include differs from ordered -f merge:\ngot: %s\nwant: %s", actual, expected)
	}
}
