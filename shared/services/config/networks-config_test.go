package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rocket-pool/smartnode/shared/types/config"
)

func TestLoadEmbeddedOfficialNetworks(t *testing.T) {
	networks, err := LoadNetworks("")
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range []config.Network{"mainnet", "testnet", "devnet"} {
		if networks.GetNetwork(name) == nil {
			t.Fatalf("missing official network %s", name)
		}
	}

	if networks.DefaultNetwork() != "mainnet" {
		t.Fatalf("default network %s, want mainnet", networks.DefaultNetwork())
	}

	mainnet := networks.GetNetwork("mainnet")
	if mainnet.ChainID != 1 {
		t.Fatalf("mainnet chainID %d", mainnet.ChainID)
	}
	if mainnet.Addresses.Storage != "0x1d8f8f00cfa6758d7bE78336684788Fb0ee0Fa46" {
		t.Fatalf("mainnet storage %s", mainnet.Addresses.Storage)
	}
	if mainnet.ClientTagSet != config.ClientTagSetProduction {
		t.Fatalf("mainnet clientTagSet %s", mainnet.ClientTagSet)
	}
	if !mainnet.IsProduction || !mainnet.SupportsMevBoost {
		t.Fatal("mainnet should be production with MEV")
	}

	testnet := networks.GetNetwork("testnet")
	if testnet.ChainID != 560048 {
		t.Fatalf("testnet chainID %d", testnet.ChainID)
	}
	if testnet.BeaconNetwork != "hoodi" {
		t.Fatalf("testnet beaconNetwork %s", testnet.BeaconNetwork)
	}

	devnet := networks.GetNetwork("devnet")
	if devnet.CustomChainConfigDir != "devnet" {
		t.Fatalf("devnet customChainConfigDir %s", devnet.CustomChainConfigDir)
	}
	if !devnet.AllowNonBlsWithdrawalCredentials {
		t.Fatal("devnet should allow non-BLS withdrawal credentials")
	}
	if got := networks.ByChainID(3151908); got == nil || got.Name != "devnet" {
		t.Fatalf("ByChainID kurtosis: %+v", got)
	}
}

func TestNetworksExtraAppendAndOverride(t *testing.T) {
	dir := t.TempDir()
	extra := []byte(`
version: 1
networks:
  - name: mainnet
    label: Forked Mainnet
    description: Local fork
    chainID: 1
    beaconNetwork: mainnet
    clientTagSet: production
    addresses:
      storage: "0x1111111111111111111111111111111111111111"
  - name: localnet
    label: Localnet
    description: e2e
    chainID: 12345
    beaconNetwork: hoodi
    clientTagSet: test
    addresses:
      storage: "0x2222222222222222222222222222222222222222"
`)
	if err := os.WriteFile(filepath.Join(dir, extraNetworksConfigPath), extra, 0644); err != nil {
		t.Fatal(err)
	}

	networks, err := LoadNetworks(dir)
	if err != nil {
		t.Fatal(err)
	}

	mainnet := networks.GetNetwork("mainnet")
	if mainnet == nil || mainnet.Addresses.Storage != "0x1111111111111111111111111111111111111111" {
		t.Fatalf("expected extra file to override mainnet storage, got %+v", mainnet)
	}
	if mainnet.Label != "Forked Mainnet" {
		t.Fatalf("label %s", mainnet.Label)
	}

	localnet := networks.GetNetwork("localnet")
	if localnet == nil {
		t.Fatal("expected localnet to be appended")
	}
	if localnet.ChainID != 12345 {
		t.Fatalf("localnet chainID %d", localnet.ChainID)
	}

	if networks.GetNetwork("testnet") == nil {
		t.Fatal("testnet should still be present from embed")
	}
}

func TestLoadNetworksMissingExtraOK(t *testing.T) {
	dir := t.TempDir()
	if _, err := LoadNetworks(dir); err != nil {
		t.Fatal(err)
	}
}

func TestLoadNetworksCorruptExtraErrors(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, extraNetworksConfigPath), []byte("not: yaml: ["), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadNetworks(dir); err == nil {
		t.Fatal("expected error for corrupt extra file")
	}
}

func TestValidateRejectsBadNetworks(t *testing.T) {
	cases := []string{
		"version: 1\nnetworks:\n  - name: all\n    label: x\n    description: y\n    chainID: 1\n    beaconNetwork: mainnet\n    addresses:\n      storage: \"0x1d8f8f00cfa6758d7bE78336684788Fb0ee0Fa46\"\n",
		"version: 1\nnetworks:\n  - name: x\n    label: x\n    description: y\n    chainID: 0\n    beaconNetwork: mainnet\n    addresses:\n      storage: \"0x1d8f8f00cfa6758d7bE78336684788Fb0ee0Fa46\"\n",
		"version: 1\nnetworks:\n  - name: x\n    label: x\n    description: y\n    chainID: 1\n    beaconNetwork: mainnet\n    addresses:\n      storage: \"\"\n",
		"version: 1\nnetworks:\n  - name: x\n    label: x\n    description: y\n    chainID: 1\n    clientTagSet: weird\n    beaconNetwork: mainnet\n    addresses:\n      storage: \"0x1d8f8f00cfa6758d7bE78336684788Fb0ee0Fa46\"\n",
		"version: 1\nnetworks:\n  - name: x\n    label: x\n    description: y\n    chainID: 1\n    beaconNetwork: mainnet\n    mevRelays:\n      notARelay: https://example\n    addresses:\n      storage: \"0x1d8f8f00cfa6758d7bE78336684788Fb0ee0Fa46\"\n",
		"version: 2\nnetworks:\n  - name: x\n    label: x\n    description: y\n    chainID: 1\n    beaconNetwork: mainnet\n    addresses:\n      storage: \"0x1d8f8f00cfa6758d7bE78336684788Fb0ee0Fa46\"\n",
	}
	for i, body := range cases {
		if _, err := parseNetworksYAML([]byte(body), "case", true); err == nil {
			t.Fatalf("case %d: expected validation error", i)
		}
	}
}

func TestExtraNetworkWithoutAddresses(t *testing.T) {
	dir := t.TempDir()
	extra := []byte(`
version: 1
networks:
  - name: holesky
    label: Holesky
    description: Sync-only public testnet with no Rocket Pool contracts
    chainID: 17000
    beaconNetwork: holesky
    clientTagSet: test
`)
	if err := os.WriteFile(filepath.Join(dir, extraNetworksConfigPath), extra, 0644); err != nil {
		t.Fatal(err)
	}

	networks, err := LoadNetworks(dir)
	if err != nil {
		t.Fatal(err)
	}
	holesky := networks.GetNetwork("holesky")
	if holesky == nil {
		t.Fatal("expected holesky extra network")
	}
	if holesky.Addresses.Storage != "" {
		t.Fatalf("storage should be empty, got %q", holesky.Addresses.Storage)
	}
	if holesky.BeaconNetwork != "holesky" {
		t.Fatalf("beaconNetwork %s", holesky.BeaconNetwork)
	}
}

func TestOfficialGetterLock(t *testing.T) {
	cfg := mustNewRocketPoolConfig(t, "", false)

	cfg.Smartnode.Network.Value = config.Network("mainnet")
	if got := cfg.Smartnode.GetStorageAddress(); got != "0x1d8f8f00cfa6758d7bE78336684788Fb0ee0Fa46" {
		t.Fatalf("mainnet storage %s", got)
	}
	if cfg.Smartnode.GetChainID() != 1 {
		t.Fatalf("mainnet chainID %d", cfg.Smartnode.GetChainID())
	}
	if tag, _ := cfg.Geth.ContainerTag.GetDefault(config.Network("mainnet")); tag != gethTagProd {
		t.Fatalf("mainnet geth tag %v", tag)
	}

	cfg.ChangeNetwork(config.Network("testnet"))
	if got := cfg.Smartnode.GetStorageAddress(); got != "0x594Fb75D3dc2DFa0150Ad03F99F97817747dd4E1" {
		t.Fatalf("testnet storage %s", got)
	}
	if cfg.Smartnode.GetChainID() != 560048 {
		t.Fatalf("testnet chainID %d", cfg.Smartnode.GetChainID())
	}

	cfg.ChangeNetwork(config.Network("devnet"))
	if got := cfg.Smartnode.GetStorageAddress(); got != "0xb4B46bdAA835F8E4b4d8e208B6559cD267851051" {
		t.Fatalf("devnet storage %s", got)
	}
	if !cfg.GetNetworkInfo().AllowNonBlsWithdrawalCredentials {
		t.Fatal("devnet should allow non-BLS credentials")
	}
}

func TestUnknownNetworkDeserializeErrors(t *testing.T) {
	cfg := mustNewRocketPoolConfig(t, "", false)
	settings := cfg.Serialize()
	settings["smartnode"]["network"] = "not-a-network"
	if err := cfg.Deserialize(settings); err == nil {
		t.Fatal("expected error for unknown network")
	}
}

func TestDiskDefaultReplacesEmbed(t *testing.T) {
	dir := t.TempDir()
	disk := []byte(`
version: 1
networks:
  - name: mainnet
    label: Disk Mainnet
    description: From disk
    default: true
    chainID: 1
    beaconNetwork: mainnet
    clientTagSet: production
    addresses:
      storage: "0x3333333333333333333333333333333333333333"
`)
	if err := os.WriteFile(filepath.Join(dir, defaultNetworksConfigPath), disk, 0644); err != nil {
		t.Fatal(err)
	}
	networks, err := LoadNetworks(dir)
	if err != nil {
		t.Fatal(err)
	}
	if networks.GetNetwork("testnet") != nil {
		t.Fatal("on-disk default should replace embed entirely")
	}
	if networks.GetNetwork("mainnet").Addresses.Storage != "0x3333333333333333333333333333333333333333" {
		t.Fatal("disk default not used")
	}
}
