package eth2

import (
	"fmt"
	"io"
	"strings"

	"github.com/rocket-pool/smartnode/shared/types/eth2/fork/deneb"
	"github.com/rocket-pool/smartnode/shared/types/eth2/fork/electra"
	"github.com/rocket-pool/smartnode/shared/types/eth2/fork/fulu"
	"github.com/rocket-pool/smartnode/shared/types/eth2/generic"
)

// State type assertions
var _ BeaconState = &electra.BeaconState{}
var _ BeaconState = &fulu.BeaconState{}

// Block type assertions
var _ SignedBeaconBlock = &deneb.SignedBeaconBlock{}
var _ SignedBeaconBlock = &electra.SignedBeaconBlock{}
var _ SignedBeaconBlock = &fulu.SignedBeaconBlock{}

type BeaconState interface {
	GetSlot() uint64
	ValidatorAndSlotProof(validatorIndex uint64) (validatorProof [][]byte, slotProof [][]byte, err error)
	HistoricalSummaryProof(slot uint64, capellaOffset uint64) ([][]byte, error)
	HistoricalSummaryBlockRootProof(slot int) ([][]byte, error)
	HistoricalSummaryStateRootProof(slot int) ([][]byte, error)
	BlockRootProof(slot uint64) ([][]byte, error)
	StateRootProof(slot uint64) ([][]byte, error)
	BlockHeaderProof() ([][]byte, error)
	GetValidators() []*generic.Validator
	GetPreviousEpochParticipation() []byte
	// PreviousEpochParticipationChunkProof proves the previous_epoch_participation
	// chunk containing validatorIndex's flags up to the state root (no
	// block-header cap). chunk is the 32-byte merkle leaf; the validator's flags
	// are at byte validatorIndex % 32 within it
	PreviousEpochParticipationChunkProof(validatorIndex uint64) (chunk [32]byte, participationProofBytes [][]byte, err error)
}

type SignedBeaconBlock interface {
	ProveWithdrawal(indexInWithdrawalsArray uint64) ([][]byte, error)
	HasExecutionPayload() bool
	Withdrawals() []*generic.Withdrawal
}

func NewBeaconState(data io.ReadCloser, fork string) (BeaconState, error) {
	fork = strings.ToLower(fork)

	defer func() {
		_ = data.Close()
	}()

	dataBytes, err := io.ReadAll(data)
	if err != nil {
		return nil, err
	}

	switch fork {
	case "electra":
		out := &electra.BeaconState{}
		err := out.UnmarshalSSZ(dataBytes)
		if err != nil {
			return nil, err
		}
		return out, nil
	case "fulu":
		out := &fulu.BeaconState{}
		err := out.UnmarshalSSZ(dataBytes)
		if err != nil {
			return nil, err
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported fork: %s", fork)
	}
}

func NewSignedBeaconBlock(data io.ReadCloser, fork string) (SignedBeaconBlock, error) {
	fork = strings.ToLower(fork)

	defer func() {
		_ = data.Close()
	}()

	dataBytes, err := io.ReadAll(data)
	if err != nil {
		return nil, err
	}

	switch fork {
	case "deneb":
		out := &deneb.SignedBeaconBlock{}
		err := out.UnmarshalSSZ(dataBytes)
		if err != nil {
			return nil, err
		}
		return out, nil
	case "electra":
		out := &electra.SignedBeaconBlock{}
		err := out.UnmarshalSSZ(dataBytes)
		if err != nil {
			return nil, err
		}
		return out, nil
	case "fulu":
		out := &fulu.SignedBeaconBlock{}
		err := out.UnmarshalSSZ(dataBytes)
		if err != nil {
			return nil, err
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported fork: %s", fork)
	}
}
