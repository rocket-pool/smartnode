package pdao

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func TestExitDeficitCommandToBatchJSON(t *testing.T) {
	for _, tt := range []struct {
		name, command, value, want string
		raw                        bool
	}{
		{"ETH", "exit-deficit", "0.2", "200000000000000000", false},
		{"wei", "exit-deficit", "200000000000000001", "200000000000000001", true},
		{"alias", "ed", "1", "1", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "batch.json")
			app := &cli.Command{Name: "rocketpool"}
			RegisterCommands(app, "pdao", nil)
			args := []string{"rocketpool", "pdao", "propose", "setting", "megapool", tt.command, "--yes", "--to-json", path}
			if tt.raw {
				args = append(args, "--raw")
			}
			args = append(args, tt.value)
			if err := app.Run(context.Background(), args); err != nil {
				t.Fatal(err)
			}
			settings, err := readBatchSettingsFile(path)
			if err != nil {
				t.Fatal(err)
			}
			want := api.PDAOBatchSetting{
				Contract: protocol.MegapoolSettingsContractName,
				Setting:  protocol.MegapoolExitDeficitPath,
				Type:     protocol.ProposalSettingTypeNameUint256,
				Value:    tt.want,
			}
			if len(settings) != 1 || settings[0] != want {
				t.Fatalf("got %+v, want %+v", settings, want)
			}
		})
	}
}

func TestWriteSettingToBatchJSON_CreateAndAppend(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "batch.json")

	err := writeSettingToBatchJSON(path, protocol.AuctionSettingsContractName, protocol.CreateLotEnabledSettingPath, "true")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	err = writeSettingToBatchJSON(path, protocol.DepositSettingsContractName, protocol.MinimumDepositSettingPath, "1000000000000000000")
	if err != nil {
		t.Fatalf("append: %v", err)
	}

	settings, err := readBatchSettingsFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(settings) != 2 {
		t.Fatalf("got %d settings, want 2", len(settings))
	}
	if settings[0].Setting != protocol.CreateLotEnabledSettingPath || settings[0].Type != protocol.ProposalSettingTypeNameBool {
		t.Fatalf("unexpected first setting: %+v", settings[0])
	}
	if settings[1].Setting != protocol.MinimumDepositSettingPath || settings[1].Type != protocol.ProposalSettingTypeNameUint256 {
		t.Fatalf("unexpected second setting: %+v", settings[1])
	}
}

func TestWriteSettingToBatchJSON_ReplaceDuplicate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "batch.json")

	if err := writeSettingToBatchJSON(path, protocol.AuctionSettingsContractName, protocol.CreateLotEnabledSettingPath, "true"); err != nil {
		t.Fatal(err)
	}
	if err := writeSettingToBatchJSON(path, protocol.AuctionSettingsContractName, protocol.CreateLotEnabledSettingPath, "false"); err != nil {
		t.Fatal(err)
	}

	settings, err := readBatchSettingsFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(settings) != 1 {
		t.Fatalf("got %d settings, want 1", len(settings))
	}
	if settings[0].Value != "false" {
		t.Fatalf("got value %q, want false", settings[0].Value)
	}
}

func TestWriteSettingToBatchJSON_RejectsUnsupportedSettings(t *testing.T) {
	for _, tt := range []struct {
		name, contract, setting, value string
		wantErr                        error
	}{
		{"address list", protocol.NetworkSettingsContractName, protocol.NetworkAllowListedControllersPath, "0x1", protocol.ErrUnsupportedBatchSetting},
		{"removed prestake period", protocol.MegapoolSettingsContractName, "prestake.challenge.period", "36", protocol.ErrUnknownPDAOSetting},
	} {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "batch.json")
			err := writeSettingToBatchJSON(path, tt.contract, tt.setting, tt.value)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("got %v, want %v", err, tt.wantErr)
			}
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Fatalf("rejected setting created an output file: %v", err)
			}
		})
	}
}

func TestReadBatchSettingsFile_MissingRequired(t *testing.T) {
	_, err := readBatchSettingsFile(filepath.Join(t.TempDir(), "missing.json"))
	if err == nil {
		t.Fatal("expected missing file error")
	}
}

func TestReadBatchSettingsFile_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte("{not-an-array}"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := readBatchSettingsFile(path); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestWriteBatchSettingsFile_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "roundtrip.json")

	want := []api.PDAOBatchSetting{
		{Contract: "a", Setting: "b", Type: "bool", Value: "true"},
	}
	if err := writeBatchSettingsFile(path, want); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got []api.PDAOBatchSetting
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != want[0] {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}
