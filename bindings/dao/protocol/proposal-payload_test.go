package protocol

import (
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"

	"github.com/rocket-pool/smartnode/bindings/types"
)

const proposalSettingMultiABI = `[{
  "name": "proposalSettingMulti",
  "type": "function",
  "stateMutability": "nonpayable",
  "inputs": [
    {"name": "_settingContractNames", "type": "string[]"},
    {"name": "_settingPaths", "type": "string[]"},
    {"name": "_types", "type": "uint8[]"},
    {"name": "_data", "type": "bytes[]"}
  ],
  "outputs": []
}]`

func TestDecodeAndFormatProposalSettingMulti(t *testing.T) {
	parsed, err := abi.JSON(strings.NewReader(proposalSettingMultiABI))
	if err != nil {
		t.Fatalf("parse ABI: %v", err)
	}

	amount := big.NewInt(0).Mul(big.NewInt(1e18), big.NewInt(1))
	address := common.HexToAddress("0x1234567890123456789012345678901234567890")
	encoded := [][]byte{
		math.PaddedBigBytes(common.Big1, 32),
		math.U256Bytes(new(big.Int).Set(amount)),
		common.LeftPadBytes(address.Bytes(), 32),
	}

	payload, err := parsed.Pack(
		"proposalSettingMulti",
		[]string{"rocketDAOProtocolSettingsAuction", "rocketDAOProtocolSettingsDeposit", "rocketDAOProtocolSettingsNetwork"},
		[]string{"auction.lot.create.enabled", "deposit.minimum", "network.node.fee.target"},
		[]uint8{
			uint8(types.ProposalSettingType_Bool),
			uint8(types.ProposalSettingType_Uint256),
			uint8(types.ProposalSettingType_Address),
		},
		encoded,
	)
	if err != nil {
		t.Fatalf("pack: %v", err)
	}

	method, err := parsed.MethodById(payload)
	if err != nil {
		t.Fatalf("method: %v", err)
	}
	args, err := method.Inputs.UnpackValues(payload[4:])
	if err != nil {
		t.Fatalf("unpack: %v", err)
	}

	settings, err := decodeProposalSettingMultiArgs(args)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(settings) != 3 {
		t.Fatalf("got %d settings, want 3", len(settings))
	}
	if settings[0].Path != "auction.lot.create.enabled" || settings[0].Value != "true" || settings[0].Type != types.ProposalSettingType_Bool {
		t.Fatalf("bool setting = %+v", settings[0])
	}
	if settings[1].Path != "deposit.minimum" || settings[1].Value != amount.String() || settings[1].Type != types.ProposalSettingType_Uint256 {
		t.Fatalf("uint setting = %+v", settings[1])
	}
	if settings[2].Path != "network.node.fee.target" || settings[2].Value != address.Hex() || settings[2].Type != types.ProposalSettingType_Address {
		t.Fatalf("address setting = %+v", settings[2])
	}

	formatted := FormatProposalSettingMulti(settings)
	if !strings.Contains(formatted, "auction.lot.create.enabled=true") {
		t.Fatalf("formatted missing bool setting: %s", formatted)
	}
	if !strings.Contains(formatted, "deposit.minimum="+amount.String()) {
		t.Fatalf("formatted missing uint setting: %s", formatted)
	}
	if !strings.Contains(formatted, "network.node.fee.target="+address.Hex()) {
		t.Fatalf("formatted missing address setting: %s", formatted)
	}
	if strings.Contains(formatted, "[") || strings.Contains(formatted, "0x000000") {
		t.Fatalf("formatted still looks like raw ABI dump: %s", formatted)
	}
}

func TestDecodeProposalSettingMultiArgs_LengthMismatch(t *testing.T) {
	_, err := decodeProposalSettingMultiArgs([]any{
		[]string{"a"},
		[]string{"b", "c"},
		[]uint8{0},
		[][]byte{{1}},
	})
	if err == nil {
		t.Fatal("expected length mismatch error")
	}
}

func TestProposalSettingTypeString(t *testing.T) {
	if types.ProposalSettingType_Bool.String() != "bool" {
		t.Fatalf("bool string = %s", types.ProposalSettingType_Bool)
	}
	if types.ProposalSettingType_Uint256.String() != "uint256" {
		t.Fatalf("uint string = %s", types.ProposalSettingType_Uint256)
	}
	if types.ProposalSettingType_Address.String() != "address" {
		t.Fatalf("address string = %s", types.ProposalSettingType_Address)
	}
}
