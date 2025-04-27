package electra

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"sync/atomic"

	"github.com/rocket-pool/smartnode/shared/types/eth2/generic"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

const beaconStateChunkCeil uint64 = 64

// Taken from https://github.com/OffchainLabs/prysm/blob/a0071826c5daf7dc3a6e76874fdaa76481a3c665/proto/prysm/v1alpha1/beacon_state.pb.go#L1955
// Unexported fields stripped, as well as proto-related field tags. JSON and ssz-size tags are preserved, and nested types are replaced with local copies as well.
type BeaconState struct {
	GenesisTime                  uint64                          `json:"genesis_time"`
	GenesisValidatorsRoot        []byte                          `json:"genesis_validators_root" ssz-size:"32"`
	Slot                         uint64                          `json:"slot"`
	Fork                         *generic.Fork                   `json:"fork"`
	LatestBlockHeader            *generic.BeaconBlockHeader      `json:"latest_block_header"`
	BlockRoots                   [8192][32]byte                  `json:"block_roots" ssz-size:"8192,32"`
	StateRoots                   [8192][32]byte                  `json:"state_roots" ssz-size:"8192,32"`
	HistoricalRoots              [][]byte                        `json:"historical_roots" ssz-max:"16777216" ssz-size:"?,32"`
	Eth1Data                     *generic.Eth1Data               `json:"eth1_data"`
	Eth1DataVotes                []*generic.Eth1Data             `json:"eth1_data_votes" ssz-max:"2048"`
	Eth1DepositIndex             uint64                          `json:"eth1_deposit_index"`
	Validators                   []*generic.Validator            `json:"validators" ssz-max:"1099511627776"`
	Balances                     []uint64                        `json:"balances" ssz-max:"1099511627776"`
	RandaoMixes                  [][]byte                        `json:"randao_mixes" ssz-size:"65536,32"`
	Slashings                    []uint64                        `json:"slashings" ssz-size:"8192"`
	PreviousEpochParticipation   []byte                          `json:"previous_epoch_participation" ssz-max:"1099511627776"`
	CurrentEpochParticipation    []byte                          `json:"current_epoch_participation" ssz-max:"1099511627776"`
	JustificationBits            [1]byte                         `json:"justification_bits" ssz-size:"1"`
	PreviousJustifiedCheckpoint  *generic.Checkpoint             `json:"previous_justified_checkpoint"`
	CurrentJustifiedCheckpoint   *generic.Checkpoint             `json:"current_justified_checkpoint"`
	FinalizedCheckpoint          *generic.Checkpoint             `json:"finalized_checkpoint"`
	InactivityScores             []uint64                        `json:"inactivity_scores" ssz-max:"1099511627776"`
	CurrentSyncCommittee         *generic.SyncCommittee          `json:"current_sync_committee"`
	NextSyncCommittee            *generic.SyncCommittee          `json:"next_sync_committee"`
	LatestExecutionPayloadHeader *generic.ExecutionPayloadHeader `json:"latest_execution_payload_header"`
	NextWithdrawalIndex          uint64                          `json:"next_withdrawal_index"`
	NextWithdrawalValidatorIndex uint64                          `json:"next_withdrawal_validator_index"`
	HistoricalSummaries          []*generic.HistoricalSummary    `json:"historical_summaries" ssz-max:"16777216"`

	// New in Electra
	DepositRequestsStartIndex     uint64                              `json:"deposit_requests_start_index"`
	DepositBalanceToConsume       uint64                              `json:"deposit_balance_to_consume"`
	ExitBalanceToConsume          uint64                              `json:"exit_balance_to_consume"`
	EarliestExitEpoch             uint64                              `json:"earliest_exit_epoch"`
	ConsolidationBalanceToConsume uint64                              `json:"consolidation_balance_to_consume"`
	EarliestConsolidationEpoch    uint64                              `json:"earliest_consolidation_epoch"`
	PendingDeposits               []*generic.PendingDeposit           `json:"pending_deposits,omitempty" ssz-max:"134217728"`
	PendingPartialWithdrawals     []*generic.PendingPartialWithdrawal `json:"pending_partial_withdrawals,omitempty" ssz-max:"134217728"`
	PendingConsolidations         []*generic.PendingConsolidation     `json:"pending_consolidations,omitempty" ssz-max:"262144"`
}

var beaconStateChunkSize atomic.Uint64

func getStateChunkSize() uint64 {
	// Use a static value to avoid multiple reflection calls
	storedChunkSize := beaconStateChunkSize.Load()
	if storedChunkSize == 0 {
		s := reflect.TypeOf(BeaconState{}).NumField()
		beaconStateChunkSize.Store(uint64(s))
		storedChunkSize = uint64(s)
	}
	return storedChunkSize
}

func GetGeneralizedIndexForValidators() uint64 {
	// There's 28 fields, so rounding up to the next power of two is 32, a left-aligned node
	// BeaconStateValidatorsIndex is the 11th field, so its generalized index is 32 + 11 = 43
	return math.GetPowerOfTwoCeil(getStateChunkSize()) + generic.BeaconStateValidatorsIndex
}

func (state *BeaconState) validatorStateProof(index uint64) ([][]byte, error) {

	// Convert the state to a proof tree
	root, err := state.GetTree()
	if err != nil {
		return nil, fmt.Errorf("could not get state tree: %w", err)
	}

	// Find the validator's generalized index
	generalizedIndex := generic.GetGeneralizedIndexForValidator(index, GetGeneralizedIndexForValidators())

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
func (state *BeaconState) ValidatorWithdrawableEpochProof(index uint64) ([][]byte, error) {

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

func (state *BeaconState) blockHeaderToStateProof(blockHeader *generic.BeaconBlockHeader) ([][]byte, error) {
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
func (state *BeaconState) ValidatorCredentialsProof(index uint64) ([][]byte, error) {

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

func (state *BeaconState) HistoricalSummaryProof(slot uint64) ([][]byte, error) {
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
	gid = gid*beaconStateChunkCeil + generic.BeaconStateHistoricalSummariesFieldIndex
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

func (state *BeaconState) HistoricalSummaryBlockRootProof(slot int) ([][]byte, error) {
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

func (state *BeaconState) BlockRootProof(slot uint64) ([][]byte, error) {
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
	gid = gid*beaconStateChunkCeil + generic.BeaconStateBlockRootsFieldIndex

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

func (state *BeaconState) GetValidators() []*generic.Validator {
	return state.Validators
}

func (state *BeaconState) GetSlot() uint64 {
	return state.Slot
}
