package pdao

import (
	"errors"
	"math/big"
	"testing"

	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func TestDecodeBatchSettings(t *testing.T) {
	entries := []api.PDAOBatchSetting{
		{
			Contract: protocol.AuctionSettingsContractName,
			Setting:  protocol.CreateLotEnabledSettingPath,
			Type:     protocol.ProposalSettingTypeNameBool,
			Value:    "true",
		},
		{
			Contract: protocol.DepositSettingsContractName,
			Setting:  protocol.MinimumDepositSettingPath,
			Value:    "1000000000000000000",
		},
	}

	decoded, err := decodeBatchSettings(entries, "")
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(decoded.values) != 2 {
		t.Fatalf("got %d values, want 2", len(decoded.values))
	}
	if decoded.settingTypes[0] != types.ProposalSettingType_Bool {
		t.Fatalf("first type = %v", decoded.settingTypes[0])
	}
	if decoded.values[0] != true {
		t.Fatalf("first value = %v", decoded.values[0])
	}
	amount, ok := decoded.values[1].(*big.Int)
	if !ok || amount.String() != "1000000000000000000" {
		t.Fatalf("second value = %v", decoded.values[1])
	}
	if decoded.message != "set auction.lot.create.enabled, deposit.minimum" {
		t.Fatalf("message = %q", decoded.message)
	}
}

func TestDecodeBatchSettings_CustomMessage(t *testing.T) {
	entries := []api.PDAOBatchSetting{
		{
			Contract: protocol.AuctionSettingsContractName,
			Setting:  protocol.CreateLotEnabledSettingPath,
			Value:    "false",
		},
	}
	decoded, err := decodeBatchSettings(entries, "update auction settings")
	if err != nil {
		t.Fatal(err)
	}
	if decoded.message != "update auction settings" {
		t.Fatalf("message = %q", decoded.message)
	}
}

func TestDecodeBatchSettings_Empty(t *testing.T) {
	_, err := decodeBatchSettings(nil, "")
	if !errors.Is(err, protocol.ErrEmptyBatchSettings) {
		t.Fatalf("got %v, want %v", err, protocol.ErrEmptyBatchSettings)
	}
}

func TestDecodeBatchSettings_Duplicate(t *testing.T) {
	entries := []api.PDAOBatchSetting{
		{Contract: protocol.AuctionSettingsContractName, Setting: protocol.CreateLotEnabledSettingPath, Value: "true"},
		{Contract: protocol.AuctionSettingsContractName, Setting: protocol.CreateLotEnabledSettingPath, Value: "false"},
	}
	_, err := decodeBatchSettings(entries, "")
	if !errors.Is(err, protocol.ErrDuplicateBatchSetting) {
		t.Fatalf("got %v, want %v", err, protocol.ErrDuplicateBatchSetting)
	}
}

func TestDecodeBatchSettings_TypeMismatch(t *testing.T) {
	entries := []api.PDAOBatchSetting{
		{
			Contract: protocol.AuctionSettingsContractName,
			Setting:  protocol.CreateLotEnabledSettingPath,
			Type:     protocol.ProposalSettingTypeNameUint256,
			Value:    "true",
		},
	}
	_, err := decodeBatchSettings(entries, "")
	if err == nil {
		t.Fatal("expected type mismatch error")
	}
}

func TestDecodeBatchSettings_AddressListRejected(t *testing.T) {
	entries := []api.PDAOBatchSetting{
		{
			Contract: protocol.NetworkSettingsContractName,
			Setting:  protocol.NetworkAllowListedControllersPath,
			Value:    "0x0000000000000000000000000000000000000001",
		},
	}
	_, err := decodeBatchSettings(entries, "")
	if !errors.Is(err, protocol.ErrUnsupportedBatchSetting) {
		t.Fatalf("got %v, want %v", err, protocol.ErrUnsupportedBatchSetting)
	}
}
