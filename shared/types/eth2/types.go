package eth2

import (
	"fmt"
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
	ValidatorProof(index uint64) ([][]byte, error)
	SlotProof(slot uint64) ([][]byte, error)
	HistoricalSummaryProof(slot uint64) ([][]byte, error)
	HistoricalSummaryBlockRootProof(slot int) ([][]byte, error)
	BlockRootProof(slot uint64) ([][]byte, error)
	BlockHeaderProof() ([][]byte, error)
	GetValidators() []*generic.Validator
}

type SignedBeaconBlock interface {
	ProveWithdrawal(indexInWithdrawalsArray uint64) ([][]byte, error)
	HasExecutionPayload() bool
	Withdrawals() []*generic.Withdrawal
}

func NewBeaconState(data []byte, fork string) (BeaconState, error) {
	fork = strings.ToLower(fork)

	switch fork {
	case "electra":
		out := &electra.BeaconState{}
		err := out.UnmarshalSSZ(data)
		if err != nil {
			return nil, err
		}
		return out, nil
	case "fulu":
		out := &fulu.BeaconState{}
		err := out.UnmarshalSSZ(data)
		if err != nil {
			return nil, err
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported fork: %s", fork)
	}
}

func NewSignedBeaconBlock(data []byte, fork string) (SignedBeaconBlock, error) {
	fork = strings.ToLower(fork)

	switch fork {
	case "deneb":
		out := &deneb.SignedBeaconBlock{}
		err := out.UnmarshalSSZ(data)
		if err != nil {
			return nil, err
		}
		return out, nil
	case "electra":
		out := &electra.SignedBeaconBlock{}
		err := out.UnmarshalSSZ(data)
		if err != nil {
			return nil, err
		}
		return out, nil
	case "fulu":
		out := &fulu.SignedBeaconBlock{}
		err := out.UnmarshalSSZ(data)
		if err != nil {
			return nil, err
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported fork: %s", fork)
	}
}
