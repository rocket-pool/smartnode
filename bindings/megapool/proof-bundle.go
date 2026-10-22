package megapool

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
)

// Proof versions accepted by BeaconStateVerifier
var (
	ValidatorProofVersion1     = big.NewInt(1)
	SlotProofVersion1          = big.NewInt(1)
	ParticipationProofVersion1 = big.NewInt(1)
	FinalBalanceProofVersion1  = big.NewInt(1)
	FinalBalanceProofVersion2  = big.NewInt(2)
)

type ValidatorProofBundleV1 struct {
	ValidatorProof ValidatorProof `json:"validatorProof"`
	SlotProof      SlotProof      `json:"slotProof"`
}

type SlotProofBundleV1 struct {
	SlotProof SlotProof `json:"slotProof"`
}

type ParticipationProofBundleV1 struct {
	ValidatorProof     ValidatorProof     `json:"validatorProof"`
	ParticipationProof ParticipationProof `json:"participationProof"`
	SlotProof          SlotProof          `json:"slotProof"`
}

type FinalBalanceProofBundleV1 struct {
	WithdrawalProof WithdrawalProof `json:"withdrawalProof"`
	ValidatorProof  ValidatorProof  `json:"validatorProof"`
	SlotProof       SlotProof       `json:"slotProof"`
}

type FinalBalanceProofBundleV2 struct {
	WithdrawalProof                  WithdrawalProof          `json:"withdrawalProof"`
	ValidatorProof                   ValidatorProof           `json:"validatorProof"`
	SlotProof                        SlotProof                `json:"slotProof"`
	PreviousNextWithdrawalIndexProof NextWithdrawalIndexProof `json:"previousNextWithdrawalIndexProof"`
}

// UsesProofBundles selects the transaction format from the deployed manager version.
// The consensus fork determines the proof witnesses independently.
func UsesProofBundles(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	manager, err := getRocketMegapoolManager(rp, opts)
	if err != nil {
		return false, err
	}
	version, err := rocketpool.GetContractVersion(rp, *manager.Address, opts)
	if err != nil {
		return false, err
	}
	switch version {
	case 1:
		return false, nil
	case 2, 3: // v1.4.1 and v1.5 use the same versioned proof-bundle interface.
		return true, nil
	default:
		return false, fmt.Errorf("unsupported rocketMegapoolManager version: %d", version)
	}
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
	participationProofComponents = []abi.ArgumentMarshaling{
		{Name: "participationSlot", Type: "uint64"},
		{Name: "validatorIndex", Type: "uint40"},
		{Name: "participationFlagsChunk", Type: "bytes32"},
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

func EncodeSlotProofBundleV1(slotProof SlotProof) ([]byte, error) {
	return encodeTuple([]abi.ArgumentMarshaling{
		{Name: "slotProof", Type: "tuple", Components: slotProofComponents},
	}, SlotProofBundleV1{SlotProof: slotProof})
}

func EncodeParticipationProofBundleV1(validatorProof ValidatorProof, participationProof ParticipationProof, slotProof SlotProof) ([]byte, error) {
	return encodeTuple([]abi.ArgumentMarshaling{
		{Name: "validatorProof", Type: "tuple", Components: validatorProofComponents},
		{Name: "participationProof", Type: "tuple", Components: participationProofComponents},
		{Name: "slotProof", Type: "tuple", Components: slotProofComponents},
	}, ParticipationProofBundleV1{ValidatorProof: validatorProof, ParticipationProof: participationProof, SlotProof: slotProof})
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
