package deneb

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"sync/atomic"

	"github.com/rocket-pool/smartnode/shared/types/eth2/generic"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

const beaconStateDenebChunkCeil uint64 = 32

// Taken from https://github.com/prysmaticlabs/prysm/blob/ac1717f1e44bd218b0bd3af0c4dec951c075f462/proto/prysm/v1alpha1/beacon_state.pb.go#L1574
// Unexported fields stripped, as well as proto-related field tags. JSON and ssz-size tags are preserved, and nested types are replaced with local copies as well.
type BeaconStateDeneb struct {
	GenesisTime                  uint64                       `json:"genesis_time"`
	GenesisValidatorsRoot        []byte                       `json:"genesis_validators_root" ssz-size:"32"`
	Slot                         uint64                       `json:"slot"`
	Fork                         *generic.Fork                `json:"fork"`
	LatestBlockHeader            *generic.BeaconBlockHeader   `json:"latest_block_header"`
	BlockRoots                   [8192][32]byte               `json:"block_roots" ssz-size:"8192,32"`
	StateRoots                   [8192][32]byte               `json:"state_roots" ssz-size:"8192,32"`
	HistoricalRoots              [][]byte                     `json:"historical_roots" ssz-max:"16777216" ssz-size:"?,32"`
	Eth1Data                     *generic.Eth1Data            `json:"eth1_data"`
	Eth1DataVotes                []*generic.Eth1Data          `json:"eth1_data_votes" ssz-max:"2048"`
	Eth1DepositIndex             uint64                       `json:"eth1_deposit_index"`
	Validators                   []*generic.Validator         `json:"validators" ssz-max:"1099511627776"`
	Balances                     []uint64                     `json:"balances" ssz-max:"1099511627776"`
	RandaoMixes                  [][]byte                     `json:"randao_mixes" ssz-size:"65536,32"`
	Slashings                    []uint64                     `json:"slashings" ssz-size:"8192"`
	PreviousEpochParticipation   []byte                       `json:"previous_epoch_participation" ssz-max:"1099511627776"`
	CurrentEpochParticipation    []byte                       `json:"current_epoch_participation" ssz-max:"1099511627776"`
	JustificationBits            [1]byte                      `json:"justification_bits" ssz-size:"1"`
	PreviousJustifiedCheckpoint  *generic.Checkpoint          `json:"previous_justified_checkpoint"`
	CurrentJustifiedCheckpoint   *generic.Checkpoint          `json:"current_justified_checkpoint"`
	FinalizedCheckpoint          *generic.Checkpoint          `json:"finalized_checkpoint"`
	InactivityScores             []uint64                     `json:"inactivity_scores" ssz-max:"1099511627776"`
	CurrentSyncCommittee         *generic.SyncCommittee       `json:"current_sync_committee"`
	NextSyncCommittee            *generic.SyncCommittee       `json:"next_sync_committee"`
	LatestExecutionPayloadHeader *ExecutionPayloadHeaderDeneb `json:"latest_execution_payload_header"`
	NextWithdrawalIndex          uint64                       `json:"next_withdrawal_index"`
	NextWithdrawalValidatorIndex uint64                       `json:"next_withdrawal_validator_index"`
	HistoricalSummaries          []*generic.HistoricalSummary `json:"historical_summaries" ssz-max:"16777216"`
}

type ExecutionPayloadHeaderDeneb struct {
	ParentHash       [32]byte        `json:"parent_hash" ssz-size:"32"`
	FeeRecipient     [20]byte        `json:"fee_recipient" ssz-size:"20"`
	StateRoot        [32]byte        `json:"state_root" ssz-size:"32"`
	ReceiptsRoot     [32]byte        `json:"receipts_root" ssz-size:"32"`
	LogsBloom        [256]byte       `json:"logs_bloom" ssz-size:"256"`
	PrevRandao       [32]byte        `json:"prev_randao" ssz-size:"32"`
	BlockNumber      uint64          `json:"block_number"`
	GasLimit         uint64          `json:"gas_limit"`
	GasUsed          uint64          `json:"gas_used"`
	Timestamp        uint64          `json:"timestamp"`
	ExtraData        []byte          `json:"extra_data" ssz-max:"32"`
	BaseFeePerGas    generic.Uint256 `json:"base_fee_per_gas" ssz-size:"32"`
	BlockHash        [32]byte        `json:"block_hash" ssz-size:"32"`
	TransactionsRoot [32]byte        `json:"transactions_root" ssz-size:"32"`
	WithdrawalRoot   [32]byte        `json:"withdrawals_root" ssz-size:"32"`
	BlobGasUsed      uint64          `json:"blob_gas_used"`
	ExcessBlobGas    uint64          `json:"excess_blob_gas"`
}

var beaconStateChunkSize atomic.Uint64

func getDenebStateChunkSize() uint64 {
	// Use a static value to avoid multiple reflection calls
	storedChunkSize := beaconStateChunkSize.Load()
	if storedChunkSize == 0 {
		s := reflect.TypeOf(BeaconStateDeneb{}).NumField()
		beaconStateChunkSize.Store(uint64(s))
		storedChunkSize = uint64(s)
	}
	return storedChunkSize
}

func GetDenebGeneralizedIndexForValidators() uint64 {
	// There's 28 fields, so rounding up to the next power of two is 32, a left-aligned node
	// BeaconStateDenebValidatorsIndex is the 11th field, so its generalized index is 32 + 11 = 43
	return math.GetPowerOfTwoCeil(getDenebStateChunkSize()) + generic.BeaconStateValidatorsIndex
}

func (state *BeaconStateDeneb) validatorStateProof(index uint64) ([][]byte, error) {

	// Convert the state to a proof tree
	root, err := state.GetTree()
	if err != nil {
		return nil, fmt.Errorf("could not get state tree: %w", err)
	}

	// Find the validator's generalized index
	generalizedIndex := generic.GetGeneralizedIndexForValidator(index, GetDenebGeneralizedIndexForValidators())

	// Grab the proof for that index
	proof, err := root.Prove(int(generalizedIndex))
	if err != nil {
		return nil, fmt.Errorf("could not get proof for validator: %w", err)
	}

	// Sanity check that the proof leaf matches the expected validator
	validatorHashTreeRoot, err := state.Validators[index].HashTreeRoot()
	if err != nil {
		return nil, fmt.Errorf("could not get hash tree root for validator: %w", err)
	}
	if !bytes.Equal(proof.Leaf, validatorHashTreeRoot[:]) {
		return nil, fmt.Errorf("proof leaf does not match expected validator")
	}

	return proof.Hashes, nil

}

// ValidatorWithdrawableEpochProof computes the merkle proof for a validator's withdrawable epoch
// at a specific index in the validator registry.
func (state *BeaconStateDeneb) ValidatorWithdrawableEpochProof(index uint64) ([][]byte, error) {

	if index >= uint64(len(state.Validators)) {
		return nil, errors.New("validator index out of bounds")
	}

	// Get the validator's withdrawable epoch proof
	withdrawableEpochProof, err := state.Validators[index].ValidatorWithdrawableEpochProof()
	if err != nil {
		return nil, fmt.Errorf("could not get validator withdrawable epoch proof: %w", err)
	}

	stateProof, err := state.validatorStateProof(index)
	if err != nil {
		return nil, fmt.Errorf("could not get validator state proof: %w", err)
	}

	// The EL proves against BeaconBlockHeader root, so we need to merge the state proof with that.
	generalizedIndex := generic.BeaconBlockHeaderStateRootGeneralizedIndex
	root, err := state.LatestBlockHeader.GetTree()
	if err != nil {
		return nil, fmt.Errorf("could not get block header tree: %w", err)
	}
	blockHeaderProof, err := root.Prove(int(generalizedIndex))
	if err != nil {
		return nil, fmt.Errorf("could not get proof for block header: %w", err)
	}

	out := append(withdrawableEpochProof, stateProof...)
	out = append(out, blockHeaderProof.Hashes...)

	return out, nil
}

func (state *BeaconStateDeneb) blockHeaderToStateProof(blockHeader *generic.BeaconBlockHeader) ([][]byte, error) {
	generalizedIndex := generic.BeaconBlockHeaderStateRootGeneralizedIndex
	root, err := blockHeader.GetTree()
	if err != nil {
		return nil, fmt.Errorf("could not get block header tree: %w", err)
	}
	blockHeaderProof, err := root.Prove(int(generalizedIndex))
	if err != nil {
		return nil, fmt.Errorf("could not get proof for block header: %w", err)
	}
	return blockHeaderProof.Hashes, nil
}

// ValidatorCredentialsProof computes the merkle proof for a validator's credentials
// at a specific index in the validator registry.
func (state *BeaconStateDeneb) ValidatorCredentialsProof(index uint64) ([][]byte, error) {

	if index >= uint64(len(state.Validators)) {
		return nil, errors.New("validator index out of bounds")
	}

	// Get the validator's credentials proof
	credentialsProof, err := state.Validators[index].ValidatorCredentialsPubkeyProof()
	if err != nil {
		return nil, fmt.Errorf("could not get validator credentials proof: %w", err)
	}

	stateProof, err := state.validatorStateProof(index)
	if err != nil {
		return nil, fmt.Errorf("could not get validator state proof: %w", err)
	}

	// The EL proves against BeaconBlockHeader root, so we need to merge the state proof with that.
	blockHeaderProof, err := state.blockHeaderToStateProof(state.LatestBlockHeader)
	if err != nil {
		return nil, fmt.Errorf("could not get block header proof: %w", err)
	}

	out := append(credentialsProof, stateProof...)
	out = append(out, blockHeaderProof...)

	return out, nil
}

func (state *BeaconStateDeneb) HistoricalSummaryProof(slot uint64) ([][]byte, error) {
	isHistorical := slot+generic.SlotsPerHistoricalRoot <= state.Slot
	if !isHistorical {
		return nil, fmt.Errorf("slot %d is less than %d slots in the past from the state at slot %d, you must build a proof from the block_roots field instead", slot, generic.SlotsPerHistoricalRoot, state.Slot)
	}
	tree, err := state.GetTree()
	if err != nil {
		return nil, fmt.Errorf("could not get state tree: %w", err)
	}

	// Navigate to the historical_summaries
	gid := uint64(1)
	gid = gid*beaconStateDenebChunkCeil + generic.BeaconStateHistoricalSummariesFieldIndex
	// Navigate into the historical summaries vector.
	arrayIndex := (slot / generic.SlotsPerHistoricalRoot)
	gid = gid*2*generic.BeaconStateHistoricalSummariesMaxLength + arrayIndex

	proof, err := tree.Prove(int(gid))
	if err != nil {
		return nil, fmt.Errorf("could not get proof for historical block root: %w", err)
	}

	// The EL proves against BeaconBlockHeader root, so we need to merge the state proof with that.
	blockHeaderProof, err := state.blockHeaderToStateProof(state.LatestBlockHeader)
	if err != nil {
		return nil, fmt.Errorf("could not get block header proof: %w", err)
	}
	return append(proof.Hashes, blockHeaderProof...), nil
}

func (state *BeaconStateDeneb) HistoricalSummaryBlockRootProof(slot int) ([][]byte, error) {
	// If the state isn't aligned at the end of an 8192 slot era, throw an error
	if state.Slot%generic.SlotsPerHistoricalRoot != generic.SlotsPerHistoricalRoot-1 {
		return nil, fmt.Errorf("state is not aligned at the end of an 8192 slot era")
	}

	if slot < int(state.Slot)-int(generic.SlotsPerHistoricalRoot)-1 || slot+int(generic.SlotsPerHistoricalRoot)-1 >= int(state.Slot) {
		return nil, fmt.Errorf("slot %d is out of range for historical summary proof", slot)
	}

	hsls := generic.HistoricalSummaryLists{
		BlockRoots: state.BlockRoots,
		StateRoots: state.StateRoots,
	}

	idx := slot % int(generic.SlotsPerHistoricalRoot)
	tree, err := hsls.GetTree()
	if err != nil {
		return nil, fmt.Errorf("could not get historical summary lists tree: %w", err)
	}

	gid := uint64(1)
	gid = gid * 2                              // Now at block_roots
	gid = gid * generic.SlotsPerHistoricalRoot // Now at the first block_root
	gid = gid + uint64(idx)                    // Now at the correct block_root

	proof, err := tree.Prove(int(gid))
	if err != nil {
		return nil, fmt.Errorf("could not get proof for historical summary: %w", err)
	}

	return proof.Hashes, nil
}

func (state *BeaconStateDeneb) BlockRootProof(slot uint64) ([][]byte, error) {
	isHistorical := slot+generic.SlotsPerHistoricalRoot <= state.Slot
	if isHistorical {
		return nil, fmt.Errorf("slot %d is more than %d slots in the past from the state at slot %d, you must build a proof from the historical_summaries instead", slot, generic.SlotsPerHistoricalRoot, state.Slot)
	}

	tree, err := state.GetTree()
	if err != nil {
		return nil, fmt.Errorf("could not get state tree: %w", err)
	}

	gid := uint64(1)

	// Navigate to the block_roots
	gid = gid*beaconStateDenebChunkCeil + generic.BeaconStateBlockRootsFieldIndex

	// We're now at the block_roots vector, which is the root of a slotsPerHistoricalRoot slots vector.
	// The index we care about is given by slot % slotsPerHistoricalRoot.
	gid = gid*generic.BeaconStateBlockRootsMaxLength + (slot % generic.SlotsPerHistoricalRoot)

	proof, err := tree.Prove(int(gid))
	if err != nil {
		return nil, fmt.Errorf("could not get proof for block root: %w", err)
	}

	// Finally, prove from the block header to the state root.
	blockHeaderProof, err := state.blockHeaderToStateProof(state.LatestBlockHeader)
	if err != nil {
		return nil, fmt.Errorf("could not get block header proof: %w", err)
	}

	return append(proof.Hashes, blockHeaderProof...), nil
}

func (state *BeaconStateDeneb) GetValidators() []*generic.Validator {
	return state.Validators
}

func (state *BeaconStateDeneb) GetSlot() uint64 {
	return state.Slot
}
