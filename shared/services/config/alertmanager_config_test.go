package config

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

// installAlertingTemplates copies the alerting template files from the source
// tree into dst, mimicking what `rocketpool service install` does at runtime.
// When `go test` runs, the working directory is the package directory
// (shared/services/config), so we walk up three levels to reach the repo root.
func installAlertingTemplates(t *testing.T, dst string) {
	t.Helper()

	repoRoot := filepath.Join("..", "..", "..")
	srcBase := filepath.Join(repoRoot, "shared", "services", "rocketpool", "assets", "install", "alerting")
	dstBase := filepath.Join(dst, "alerting")

	err := filepath.Walk(srcBase, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcBase, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dstBase, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		return copyFile(target, path)
	})
	if err != nil {
		t.Fatalf("failed to install alerting templates: %v", err)
	}
}

func copyFile(dst, src string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// TestUpdateConfigurationFiles_HighStorageRules verifies that high-storage.yml
// is created when at least one client is locally managed, and deleted when
// both clients are externally managed.
func TestUpdateConfigurationFiles_HighStorageRules(t *testing.T) {
	tests := []struct {
		name          string
		ecMode        cfgtypes.Mode
		ccMode        cfgtypes.Mode
		expectCreated bool
	}{
		{
			name:          "local EC + local CC creates high-storage.yml",
			ecMode:        cfgtypes.Mode_Local,
			ccMode:        cfgtypes.Mode_Local,
			expectCreated: true,
		},
		{
			name:          "external EC + external CC removes high-storage.yml",
			ecMode:        cfgtypes.Mode_External,
			ccMode:        cfgtypes.Mode_External,
			expectCreated: false,
		},
		{
			name:          "local EC + external CC creates high-storage.yml",
			ecMode:        cfgtypes.Mode_Local,
			ccMode:        cfgtypes.Mode_External,
			expectCreated: true,
		},
		{
			name:          "external EC + local CC creates high-storage.yml",
			ecMode:        cfgtypes.Mode_External,
			ccMode:        cfgtypes.Mode_Local,
			expectCreated: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Use a fresh temp dir for each sub-test so they are independent.
			testDir := t.TempDir()
			installAlertingTemplates(t, testDir)

			cfg := NewRocketPoolConfig(testDir, false)
			cfg.ExecutionClientMode.Value = tc.ecMode
			cfg.ConsensusClientMode.Value = tc.ccMode
			cfg.Alertmanager.EnableAlerting.Value = true

			if err := cfg.Alertmanager.UpdateConfigurationFiles(testDir); err != nil {
				t.Fatalf("UpdateConfigurationFiles returned error: %v", err)
			}

			highStorageFile := filepath.Join(testDir, HighStorageRulesConfigFile)
			_, statErr := os.Stat(highStorageFile)
			exists := !os.IsNotExist(statErr)

			if exists != tc.expectCreated {
				if tc.expectCreated {
					t.Errorf("expected high-storage.yml to be created, but it was not")
				} else {
					t.Errorf("expected high-storage.yml to be deleted, but it still exists")
				}
			}
		})
	}
}
