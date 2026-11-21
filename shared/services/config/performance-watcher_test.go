package config

import (
	"strings"
	"testing"

	configtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

func TestPerformanceWatcherConfiguration(t *testing.T) {
	for _, native := range []bool{false, true} {
		cfg := mustNewRocketPoolConfig(t, t.TempDir(), native)
		if cfg.Smartnode.RunPerformanceWatcher.Value != false || cfg.Smartnode.PerformanceWatcherBackfill.Value != false || cfg.Smartnode.PerformanceWatcherBootstrapRetentionDays.Value != uint64(201) {
			t.Fatal("watcher defaults")
		}
		found := map[string]bool{}
		names := map[string]string{"runPerformanceWatcher": "Run Performance Watcher", "performanceWatcherBackfill": "Backfill Performance History"}
		for _, p := range cfg.Smartnode.GetParameters() {
			if name, ok := names[p.ID]; ok {
				found[p.ID] = true
				if p.Name != name || p.Type != configtypes.ParameterType_Bool || len(p.AffectsContainers) != 1 || p.AffectsContainers[0] != configtypes.ContainerID_Node || p.OverwriteOnUpgrade {
					t.Fatal("checkbox/restart behavior")
				}
			}
		}
		if len(found) != 2 {
			t.Fatal("watcher not exposed to configuration UI")
		}
		cfg.Smartnode.RunPerformanceWatcher.Value = true
		cfg.Smartnode.PerformanceWatcherBackfill.Value = true
		cfg.Smartnode.PerformanceWatcherBootstrapRetentionDays.Value = uint64(230)
		loaded := mustNewRocketPoolConfig(t, t.TempDir(), native)
		if err := loaded.Deserialize(cfg.Serialize()); err != nil {
			t.Fatal(err)
		}
		if loaded.Smartnode.RunPerformanceWatcher.Value != true || loaded.Smartnode.PerformanceWatcherBackfill.Value != true || loaded.Smartnode.PerformanceWatcherBootstrapRetentionDays.Value != uint64(230) {
			t.Fatal("watcher settings not preserved")
		}
		legacy := cfg.Serialize()
		delete(legacy["smartnode"], "performanceWatcherBackfill")
		if err := loaded.Deserialize(legacy); err != nil {
			t.Fatal(err)
		}
		if loaded.Smartnode.PerformanceWatcherBackfill.Value != false {
			t.Fatal("existing configurations must default to live only")
		}
		loaded.Smartnode.PerformanceWatcherBootstrapRetentionDays.Value = uint64(0)
		if !strings.Contains(strings.Join(loaded.Validate(), " "), "Retention Days must be positive") {
			t.Fatal("zero retention accepted")
		}
	}
}
