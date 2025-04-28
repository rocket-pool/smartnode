package eth2

import (
	"fmt"
	"strings"

	"github.com/rocket-pool/smartnode/shared/types/eth2/fork/deneb"
	"github.com/rocket-pool/smartnode/shared/types/eth2/fork/electra"
	"github.com/rocket-pool/smartnode/shared/types/eth2/generic"
)

// State type assertions
var _ BeaconState = &deneb.BeaconState{}
var _ BeaconState = &electra.BeaconState{}

// Block type assertions
var _ BeaconBlock = &deneb.BeaconBlock{}
var _ BeaconBlock = &electra.BeaconBlock{}

type BeaconState interface {
	GetSlot() uint64
	ValidatorWithdrawableEpochProof(index uint64) ([][]byte, error)
	ValidatorCredentialsProof(index uint64) ([][]byte, error)
	HistoricalSummaryProof(slot uint64) ([][]byte, error)
	HistoricalSummaryBlockRootProof(slot int) ([][]byte, error)
	BlockRootProof(slot uint64) ([][]byte, error)
	GetValidators() []*generic.Validator
}

type BeaconBlock interface {
	ProveWithdrawal(indexInWithdrawalsArray uint64) ([][]byte, error)
	HasExecutionPayload() bool
	Withdrawals() []*generic.Withdrawal
}

func NewBeaconState(data []byte, fork string) (BeaconState, error) {
	fork = strings.ToLower(fork)

	switch fork {
	case "deneb":
		out := &deneb.BeaconState{}
		err := out.UnmarshalSSZ(data)
		if err != nil {
			return nil, err
		}
		return out, nil
	case "electra":
		out := &electra.BeaconState{}
		err := out.UnmarshalSSZ(data)
		if err != nil {
			return nil, err
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported fork: %s", fork)
	}
}

func NewBeaconBlock(data []byte, fork string) (BeaconBlock, error) {
	fork = strings.ToLower(fork)

	switch fork {
	case "deneb":
		out := &deneb.BeaconBlock{}
		err := out.UnmarshalSSZ(data)
		if err != nil {
			return nil, err
		}
		return out, nil
	case "electra":
		out := &electra.BeaconBlock{}
		err := out.UnmarshalSSZ(data)
		if err != nil {
			return nil, err
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported fork: %s", fork)
	}
}
