package config

import (
	"reflect"
	"strings"
	"testing"

	"github.com/rocket-pool/smartnode/shared/types/config"
)

func optionValues(param *config.Parameter) []any {
	values := make([]any, len(param.Options))
	for i, option := range param.Options {
		values[i] = option.Value
	}
	return values
}

func TestPlatabergetConfiguration(t *testing.T) {
	cfg := mustNewRocketPoolConfig(t, "", false)
	cfg.ExecutionClient.Value = config.ExecutionClient_Besu
	cfg.ConsensusClient.Value = config.ConsensusClient_Prysm
	cfg.EnableCommitBoost.Value = true
	cfg.ChangeNetwork("plataberget")
	if got := cfg.Smartnode.GetChainID(); got != 7091047534 {
		t.Fatalf("chain ID: %d", got)
	}
	if cfg.Smartnode.GetBeaconNetwork() != "plataberget" || cfg.Smartnode.GetCustomChainConfigDir() != "" {
		t.Fatal("Plataberget must use a named network without custom chain files")
	}
	if cfg.ExecutionClient.Value != config.ExecutionClient_Nethermind || cfg.ConsensusClient.Value != config.ConsensusClient_Nimbus {
		t.Fatalf("unsupported selections were not replaced: %v / %v", cfg.ExecutionClient.Value, cfg.ConsensusClient.Value)
	}
	if cfg.EnableMevBoost.Value != false || cfg.EnableCommitBoost.Value != false {
		t.Fatal("PBS must be disabled")
	}
	if cfg.ConsensusCommon.CheckpointSyncProvider.Value != "https://checkpoint-sync.plataberget.ethpandaops.io" {
		t.Fatal("wrong checkpoint sync default")
	}
	if problems := cfg.Validate(); len(problems) != 0 {
		t.Fatalf("pending contracts must not prevent saving settings: %v", problems)
	}
	if err := cfg.ValidateNetworkForStart(); err == nil || !strings.Contains(err.Error(), "contracts are not configured") {
		t.Fatalf("startup should report pending contracts: %v", err)
	}
	if cfg.Smartnode.GetStorageAddress() != "" || cfg.GetNetworkInfo().Addresses.Reth != "" {
		t.Fatal("contract addresses must remain unset")
	}
	wantEC := []any{config.ExecutionClient_Nethermind}
	wantCC := []any{config.ConsensusClient_Lodestar, config.ConsensusClient_Nimbus, config.ConsensusClient_Teku}
	for _, candidate := range []*RocketPoolConfig{cfg, cfg.CreateCopy()} {
		if !reflect.DeepEqual(optionValues(&candidate.ExecutionClient), wantEC) {
			t.Fatalf("execution choices: %v", optionValues(&candidate.ExecutionClient))
		}
		for _, param := range []*config.Parameter{&candidate.ConsensusClient, &candidate.ExternalConsensusClient, &candidate.Native.ConsensusClient} {
			if !reflect.DeepEqual(optionValues(param), wantCC) {
				t.Fatalf("%s choices: %v", param.ID, optionValues(param))
			}
		}
	}
	loaded := mustNewRocketPoolConfig(t, "", false)
	if err := loaded.Deserialize(cfg.Serialize()); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(optionValues(&loaded.ConsensusClient), wantCC) || loaded.GetNetwork() != "plataberget" {
		t.Fatal("saved network restrictions were lost")
	}
	loaded.ChangeNetwork("testnet")
	if len(loaded.ExecutionClient.Options) != 5 || len(loaded.ConsensusClient.Options) != 5 || len(loaded.Native.ConsensusClient.Options) != 5 {
		t.Fatal("switching to Hoodi did not restore unrestricted choices")
	}
	if loaded.Smartnode.GetBeaconNetwork() != "hoodi" || loaded.ConsensusCommon.CheckpointSyncProvider.Value != "" {
		t.Fatal("Hoodi defaults were not restored")
	}
}

func TestPlatabergetSavedUnsupportedClients(t *testing.T) {
	for _, mode := range []string{"local", "external", "native"} {
		t.Run(mode, func(t *testing.T) {
			cfg := mustNewRocketPoolConfig(t, "", mode == "native")
			cfg.ChangeNetwork("plataberget")
			switch mode {
			case "local":
				cfg.ExecutionClient.Value = config.ExecutionClient_Geth
			case "external":
				cfg.ExecutionClientMode.Value = config.Mode_External
				cfg.ConsensusClientMode.Value = config.Mode_External
				cfg.ExternalConsensusClient.Value = config.ConsensusClient_Lighthouse
			case "native":
				cfg.Native.ConsensusClient.Value = config.ConsensusClient_Prysm
			}
			loaded := mustNewRocketPoolConfig(t, "", false)
			if err := loaded.Deserialize(cfg.Serialize()); err != nil {
				t.Fatal(err)
			}
			if err := loaded.ValidateNetworkClients(); err == nil || !strings.Contains(err.Error(), "not supported on Platåberget") {
				t.Fatalf("unsupported saved client accepted: %v", err)
			}
		})
	}
}

func TestPlatabergetSupportedClientsAndDeployment(t *testing.T) {
	cfg := mustNewRocketPoolConfig(t, "", false)
	cfg.ChangeNetwork("plataberget")
	// A deployment replaces the pending network entry; never ship this fixture address.
	network := cfg.GetNetworkInfo()
	network.Addresses.Storage = "0x1111111111111111111111111111111111111111"
	if err := cfg.ValidateNetworkForStart(); err == nil {
		t.Fatal("pending flag must still block partial deployments")
	}
	network.ContractsPending = false
	for _, cc := range []config.ConsensusClient{config.ConsensusClient_Lodestar, config.ConsensusClient_Nimbus, config.ConsensusClient_Teku} {
		cfg.ConsensusClient.Value = cc
		if err := cfg.ValidateNetworkForStart(); err != nil {
			t.Fatalf("%s: %v", cc, err)
		}
	}
	cfg.EnableMevBoost.Value = true
	if err := cfg.ValidateNetworkClients(); err == nil {
		t.Fatal("unsupported PBS configuration accepted")
	}
}

func TestNetworkClientSchemaValidation(t *testing.T) {
	for _, extra := range []string{
		"executionClients: [unknown]", "consensusClients: [unknown]",
		"executionClients: [nethermind, nethermind]", "consensusClients: [teku, teku]",
		"isProduction: true", "contractsPending: false",
	} {
		body := "version: 1\nnetworks:\n  - name: pending\n    label: Pending\n    description: Pending network\n    chainID: 123\n    beaconNetwork: pending\n    contractsPending: true\n    " + extra + "\n"
		if _, err := parseNetworksYAML([]byte(body), "test", true); err == nil {
			t.Fatalf("invalid network accepted: %s", extra)
		}
	}
}
