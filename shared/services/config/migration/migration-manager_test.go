package migration

import (
	"slices"
	"testing"
)

func TestUpdateConfigRunsAllApplicableMigrations(t *testing.T) {
	serialized := map[string]map[string]string{
		"root": {
			"version": "v1.3.1", "executionClient": "geth",
			"enableMevBoost": "true", "enableCommitBoost": "false",
		},
		"geth": {
			"p2pPort": "30305", "ethstatsLabel": "node", "ethstatsLogin": "login", "archiveMode": "true",
		},
		"nimbus": {"additionalFlags": "--log-level=debug"},
		"executionCommon": {
			"openRpcPorts": "true", "suggestedBlockGasLimit": "36000000",
		},
		"consensusCommon": {
			"openApiPort": "false", "suggestedBlockGasLimit": "36000000", "checkpointSyncUrl": "https://beaconstate.info",
		},
		"nethermind": {"pruneMemSize": "512", "fullPruneMemoryBudget": "512"},
		"smartnode": {
			"network": "mainnet", "priorityFee": "2", "minipoolStakeGasThreshold": "150", "apiPort": "9001",
		},
		"mevBoost": {"selectionMode": "relay", "bloxRouteMaxProfitEnabled": "true"},
	}
	if err := UpdateConfig(serialized); err != nil {
		t.Fatal(err)
	}
	for _, check := range []struct {
		section string
		key     string
		want    string
	}{
		{"executionCommon", "p2pPort", "30305"},
		{"executionCommon", "ethstatsLabel", "node"},
		{"executionCommon", "ethstatsLogin", "login"},
		{"nimbus", "additionalBnFlags", "--log-level=debug"},
		{"executionCommon", "openRpcPorts", "localhost"},
		{"consensusCommon", "openApiPort", "closed"},
		{"executionCommon", "suggestedBlockGasLimit", ""},
		{"consensusCommon", "suggestedBlockGasLimit", ""},
		{"executionCommon", "pruningMode", "archive"},
		{"nethermind", "pruneMemSize", ""},
		{"nethermind", "fullPruneMemoryBudget", ""},
		{"smartnode", "priorityFee", "0.01"},
		{"smartnode", "minipoolStakeGasThreshold", "20"},
		{"api", "apiPort", "9001"},
		{"mevBoost", "bloxRouteRegulatedEnabled", "true"},
		{"epbs", "bloxRouteRegulatedEnabled", "true"},
		{"epbs", "flashbotsEnabled", "false"},
	} {
		if got := serialized[check.section][check.key]; got != check.want {
			t.Errorf("%s.%s = %q, want %q", check.section, check.key, got, check.want)
		}
	}
	if !slices.Contains(checkpointSyncUrlReplacements, serialized["consensusCommon"]["checkpointSyncUrl"]) {
		t.Fatal("checkpoint URL migration was skipped")
	}
	if _, exists := serialized["smartnode"]["apiPort"]; exists {
		t.Fatal("legacy API port was not removed")
	}
	if _, exists := serialized["mevBoost"]["bloxRouteMaxProfitEnabled"]; exists {
		t.Fatal("retired relay was not migrated before initializing ePBS settings")
	}
}
