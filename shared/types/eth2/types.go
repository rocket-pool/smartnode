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
	BlockRootProof(slot uint64) ([][]byte, error)
	BlockHeaderProof() ([][]byte, error)
	GetValidators() []*generic.Validator
}

type SignedBeaconBlock interface {
	ProveWithdrawal(indexInWithdrawalsArray uint64) ([][]byte, error)
	HasExecutionPayload() bool
	Withdrawals() []*generic.Withdrawal
}

// decodeSSZ deserializes an SSZ payload into target. When the total payload
// size is known it streams directly from the reader, avoiding holding the
// whole serialized payload (~310 MB for a mainnet beacon state) in memory
// alongside the decoded struct. When the size is unknown (e.g. a chunked
// response without Content-Length) it falls back to buffering, since SSZ
// offsets cannot be interpreted without the total size.
func decodeSSZ(target any, data io.ReadCloser, size int64) error {
	defer func() {
		_ = data.Close()
	}()

	if size > 0 {
		return generic.SSZ.UnmarshalSSZReader(target, data, int(size))
	}

	dataBytes, err := io.ReadAll(data)
	if err != nil {
		return err
	}
	return generic.SSZ.UnmarshalSSZ(target, dataBytes)
}

func NewBeaconState(data io.ReadCloser, size int64, fork string) (BeaconState, error) {
	fork = strings.ToLower(fork)

	switch fork {
	case "electra":
		out := &electra.BeaconState{}
		err := decodeSSZ(out, data, size)
		if err != nil {
			return nil, err
		}
		return out, nil
	case "fulu":
		out := &fulu.BeaconState{}
		err := decodeSSZ(out, data, size)
		if err != nil {
			return nil, err
		}
		return out, nil
	default:
		_ = data.Close()
		return nil, fmt.Errorf("unsupported fork: %s", fork)
	}
}

func NewSignedBeaconBlock(data io.ReadCloser, size int64, fork string) (SignedBeaconBlock, error) {
	fork = strings.ToLower(fork)

	switch fork {
	case "deneb":
		out := &deneb.SignedBeaconBlock{}
		err := decodeSSZ(out, data, size)
		if err != nil {
			return nil, err
		}
		return out, nil
	case "electra":
		out := &electra.SignedBeaconBlock{}
		err := decodeSSZ(out, data, size)
		if err != nil {
			return nil, err
		}
		return out, nil
	case "fulu":
		out := &fulu.SignedBeaconBlock{}
		err := decodeSSZ(out, data, size)
		if err != nil {
			return nil, err
		}
		return out, nil
	default:
		_ = data.Close()
		return nil, fmt.Errorf("unsupported fork: %s", fork)
	}
}
