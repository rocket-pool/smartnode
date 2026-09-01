package megapool

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// Proof versions accepted by BeaconStateVerifier on v1.4.1-dev.
var (
	ValidatorProofVersion1    = big.NewInt(1)
	FinalBalanceProofVersion1 = big.NewInt(1)
	FinalBalanceProofVersion2 = big.NewInt(2)
)

type ValidatorProofBundleV1 struct {
	ValidatorProof ValidatorProof
	SlotProof      SlotProof
}

type FinalBalanceProofBundleV1 struct {
	WithdrawalProof WithdrawalProof
	ValidatorProof  ValidatorProof
	SlotProof       SlotProof
}

type FinalBalanceProofBundleV2 struct {
	WithdrawalProof                  WithdrawalProof
	ValidatorProof                   ValidatorProof
	SlotProof                        SlotProof
	PreviousNextWithdrawalIndexProof NextWithdrawalIndexProof
	ValidatorBalanceProof            ValidatorBalanceProof
}

var (
	validatorComponents = []abi.ArgumentMarshaling{
		{Name: "pubkey", Type: "bytes"},
		{Name: "withdrawalCredentials", Type: "bytes32"},
		{Name: "effectiveBalance", Type: "uint64"},
		{Name: "slashed", Type: "bool"},
		{Name: "activationEligibilityEpoch", Type: "uint64"},
		{Name: "activationEpoch", Type: "uint64"},
		{Name: "exitEpoch", Type: "uint64"},
		{Name: "withdrawableEpoch", Type: "uint64"},
	}
	validatorProofComponents = []abi.ArgumentMarshaling{
		{Name: "validatorIndex", Type: "uint40"},
		{Name: "validator", Type: "tuple", Components: validatorComponents},
		{Name: "witnesses", Type: "bytes32[]"},
	}
	slotProofComponents = []abi.ArgumentMarshaling{
		{Name: "slot", Type: "uint64"},
		{Name: "witnesses", Type: "bytes32[]"},
	}
	withdrawalComponents = []abi.ArgumentMarshaling{
		{Name: "index", Type: "uint64"},
		{Name: "validatorIndex", Type: "uint64"},
		{Name: "withdrawalCredentials", Type: "bytes20"},
		{Name: "amountInGwei", Type: "uint64"},
	}
	withdrawalProofComponents = []abi.ArgumentMarshaling{
		{Name: "withdrawalSlot", Type: "uint64"},
		{Name: "withdrawalNum", Type: "uint16"},
		{Name: "withdrawal", Type: "tuple", Components: withdrawalComponents},
		{Name: "witnesses", Type: "bytes32[]"},
	}
	nextWithdrawalIndexProofComponents = []abi.ArgumentMarshaling{
		{Name: "nextWithdrawalIndex", Type: "uint64"},
		{Name: "witnesses", Type: "bytes32[]"},
	}
	validatorBalanceProofComponents = []abi.ArgumentMarshaling{
		{Name: "balanceChunk", Type: "bytes32"},
		{Name: "witnesses", Type: "bytes32[]"},
	}
)

func EncodeValidatorProofBundleV1(validatorProof ValidatorProof, slotProof SlotProof) ([]byte, error) {
	return encodeTuple([]abi.ArgumentMarshaling{
		{Name: "validatorProof", Type: "tuple", Components: validatorProofComponents},
		{Name: "slotProof", Type: "tuple", Components: slotProofComponents},
	}, ValidatorProofBundleV1{
		ValidatorProof: validatorProof,
		SlotProof:      slotProof,
	})
}

func EncodeFinalBalanceProofBundleV1(withdrawalProof WithdrawalProof, validatorProof ValidatorProof, slotProof SlotProof) ([]byte, error) {
	return encodeTuple([]abi.ArgumentMarshaling{
		{Name: "withdrawalProof", Type: "tuple", Components: withdrawalProofComponents},
		{Name: "validatorProof", Type: "tuple", Components: validatorProofComponents},
		{Name: "slotProof", Type: "tuple", Components: slotProofComponents},
	}, FinalBalanceProofBundleV1{
		WithdrawalProof: withdrawalProof,
		ValidatorProof:  validatorProof,
		SlotProof:       slotProof,
	})
}

func EncodeFinalBalanceProofBundleV2(
	withdrawalProof WithdrawalProof,
	validatorProof ValidatorProof,
	slotProof SlotProof,
	previousNextWithdrawalIndexProof NextWithdrawalIndexProof,
	validatorBalanceProof ValidatorBalanceProof,
) ([]byte, error) {
	return encodeTuple([]abi.ArgumentMarshaling{
		{Name: "withdrawalProof", Type: "tuple", Components: withdrawalProofComponents},
		{Name: "validatorProof", Type: "tuple", Components: validatorProofComponents},
		{Name: "slotProof", Type: "tuple", Components: slotProofComponents},
		{Name: "previousNextWithdrawalIndexProof", Type: "tuple", Components: nextWithdrawalIndexProofComponents},
		{Name: "validatorBalanceProof", Type: "tuple", Components: validatorBalanceProofComponents},
	}, FinalBalanceProofBundleV2{
		WithdrawalProof:                  withdrawalProof,
		ValidatorProof:                   validatorProof,
		SlotProof:                        slotProof,
		PreviousNextWithdrawalIndexProof: previousNextWithdrawalIndexProof,
		ValidatorBalanceProof:            validatorBalanceProof,
	})
}

func encodeTuple(components []abi.ArgumentMarshaling, value any) ([]byte, error) {
	typ, err := abi.NewType("tuple", "", components)
	if err != nil {
		return nil, fmt.Errorf("error creating proof bundle tuple type: %w", err)
	}
	encoded, err := abi.Arguments{{Type: typ}}.Pack(value)
	if err != nil {
		return nil, fmt.Errorf("error encoding proof bundle: %w", err)
	}
	return encoded, nil
}
