package megapool

import (
	"math/big"
	"testing"
)

func sampleValidatorProof() ValidatorProof {
	return ValidatorProof{
		ValidatorIndex: big.NewInt(7),
		Validator: ProvedValidator{
			Pubkey:                     make([]byte, 48),
			WithdrawalCredentials:      [32]byte{0x01},
			EffectiveBalance:           32_000_000_000,
			Slashed:                    false,
			ActivationEligibilityEpoch: 1,
			ActivationEpoch:            2,
			ExitEpoch:                  3,
			WithdrawableEpoch:          4,
		},
		Witnesses: [][32]byte{{0x11}, {0x22}},
	}
}

func sampleSlotProof() SlotProof {
	return SlotProof{
		Slot:      99,
		Witnesses: [][32]byte{{0x33}},
	}
}

func sampleWithdrawalProof() WithdrawalProof {
	return WithdrawalProof{
		WithdrawalSlot: 50,
		WithdrawalNum:  1,
		Withdrawal: Withdrawal{
			Index:                 10,
			ValidatorIndex:        7,
			WithdrawalCredentials: [20]byte{0xaa},
			AmountInGwei:          32_000_000_000,
		},
		Witnesses: [][32]byte{{0x44}, {0x55}},
	}
}

func TestEncodeValidatorProofBundleV1(t *testing.T) {
	encoded, err := EncodeValidatorProofBundleV1(sampleValidatorProof(), sampleSlotProof())
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if len(encoded) == 0 {
		t.Fatal("expected non-empty encoding")
	}
}

func TestEncodeFinalBalanceProofBundleV1(t *testing.T) {
	encoded, err := EncodeFinalBalanceProofBundleV1(sampleWithdrawalProof(), sampleValidatorProof(), sampleSlotProof())
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if len(encoded) == 0 {
		t.Fatal("expected non-empty encoding")
	}
}

func TestEncodeFinalBalanceProofBundleV2(t *testing.T) {
	encoded, err := EncodeFinalBalanceProofBundleV2(
		sampleWithdrawalProof(),
		sampleValidatorProof(),
		sampleSlotProof(),
		NextWithdrawalIndexProof{
			NextWithdrawalIndex: 9,
			Witnesses:           [][32]byte{{0x66}},
		},
		ValidatorBalanceProof{
			BalanceChunk: [32]byte{0x77},
			Witnesses:    [][32]byte{{0x88}},
		},
	)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if len(encoded) == 0 {
		t.Fatal("expected non-empty encoding")
	}
}
