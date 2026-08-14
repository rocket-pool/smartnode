package pdao

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

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

func TestWriteSettingToBatchJSON_RejectsAddressList(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "batch.json")

	err := writeSettingToBatchJSON(path, protocol.NetworkSettingsContractName, protocol.NetworkAllowListedControllersPath, "0x1")
	if err == nil {
		t.Fatal("expected error for address list setting")
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
