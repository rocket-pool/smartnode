package fulu

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

	// New in Fulu
	ProposerLookahead []uint64 `json:"proposer_lookahead,omitempty" ssz-size:"64"`
}

var beaconStateChunkSize atomic.Uint64

func getStateChunkSize() uint64 {
	// Use a static value to avoid multiple reflection calls
	storedChunkSize := beaconStateChunkSize.Load()
	if storedChunkSize == 0 {
		s := reflect.TypeFor[BeaconState]().NumField()
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

func GetGeneralizedIndexForSlot() uint64 {
	// There's 28 fields, so rounding up to the next power of two is 32, a left-aligned node
	// BeaconStateValidatorsIndex is the 2nd field, so its generalized index is 32 + 2 = 34
	return math.GetPowerOfTwoCeil(getStateChunkSize()) + generic.BeaconStateSlotIndex
}

// ValidatorAndSlotProof produces both the validator proof and the slot proof
// for the state's current slot
func (state *BeaconState) ValidatorAndSlotProof(validatorIndex uint64) ([][]byte, [][]byte, error) {

	if validatorIndex >= uint64(len(state.Validators)) {
		return nil, nil, errors.New("validator index out of bounds")
	}

	stateTree, err := state.GetTree()
	if err != nil {
		return nil, nil, fmt.Errorf("could not get state tree: %w", err)
	}

	validatorGid := generic.GetGeneralizedIndexForValidator(validatorIndex, GetGeneralizedIndexForValidators())
	validatorStateProof, err := stateTree.Prove(int(validatorGid))
	if err != nil {
		return nil, nil, fmt.Errorf("could not get proof for validator: %w", err)
	}

	// Sanity check that the proof leaf matches the expected validator
	validatorHashTreeRoot, err := state.Validators[validatorIndex].HashTreeRoot()
	if err != nil {
		return nil, nil, fmt.Errorf("could not get hash tree root for validator: %w", err)
	}
	if !bytes.Equal(validatorStateProof.Leaf, validatorHashTreeRoot[:]) {
		return nil, nil, fmt.Errorf("proof leaf does not match expected validator")
	}

	slotStateProof, err := stateTree.Prove(int(GetGeneralizedIndexForSlot()))
	if err != nil {
		return nil, nil, fmt.Errorf("could not get proof for slot: %w", err)
	}

	// Drop the state tree before doing more work so the GC can reclaim it.
	stateTree = nil

	bhTree, err := state.LatestBlockHeader.GetTree()
	if err != nil {
		return nil, nil, fmt.Errorf("could not get block header tree: %w", err)
	}
	blockHeaderProof, err := bhTree.Prove(int(generic.BeaconBlockHeaderStateRootGeneralizedIndex))
	if err != nil {
		return nil, nil, fmt.Errorf("could not get proof for block header: %w", err)
	}

	validatorProof := make([][]byte, 0, len(validatorStateProof.Hashes)+len(blockHeaderProof.Hashes))
	validatorProof = append(validatorProof, validatorStateProof.Hashes...)
	validatorProof = append(validatorProof, blockHeaderProof.Hashes...)

	slotProof := make([][]byte, 0, len(slotStateProof.Hashes)+len(blockHeaderProof.Hashes))
	slotProof = append(slotProof, slotStateProof.Hashes...)
	slotProof = append(slotProof, blockHeaderProof.Hashes...)

	return validatorProof, slotProof, nil
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

func (state *BeaconState) HistoricalSummaryProof(slot uint64, capellaOffset uint64) ([][]byte, error) {
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
	arrayIndex := (slot / generic.SlotsPerHistoricalRoot) - capellaOffset
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
	if state.Slot%generic.SlotsPerHistoricalRoot != 0 {
		return nil, fmt.Errorf("state is not aligned at the end of an 8192 slot era")
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

	return proof.Hashes, nil
}

func (state *BeaconState) BlockHeaderProof() ([][]byte, error) {
	// Construct block header with state root
	stateRoot, err := state.HashTreeRoot()
	if err != nil {
		return nil, fmt.Errorf("could not get state root: %w", err)
	}
	latestBlockHeader := state.LatestBlockHeader
	blockHeader := *latestBlockHeader
	blockHeader.StateRoot = stateRoot[:]
	blockHeaderTree, err := blockHeader.GetTree()
	if err != nil {
		return nil, fmt.Errorf("could not get block header tree: %w", err)
	}
	blockHeaderProofResult, err := blockHeaderTree.Prove(int(generic.BeaconBlockHeaderStateRootGeneralizedIndex))
	if err != nil {
		return nil, fmt.Errorf("could not get block header proof: %w", err)
	}
	return blockHeaderProofResult.Hashes, nil

}

func (state *BeaconState) GetValidators() []*generic.Validator {
	return state.Validators
}

func (state *BeaconState) GetSlot() uint64 {
	return state.Slot
}

// GetPendingDeposits returns all entries in the state's pending_deposits queue
// whose pubkey matches the supplied one.
func (state *BeaconState) GetPendingDeposits(pubkey []byte) ([]*generic.PendingDeposit, error) {
	var deposits []*generic.PendingDeposit
	for _, deposit := range state.PendingDeposits {
		if deposit == nil {
			continue
		}
		if bytes.Equal(deposit.Pubkey, pubkey) {
			deposits = append(deposits, deposit)
		}
	}
	return deposits, nil
}

// PendingDepositProof builds a Merkle proof that the oldest pending deposit
// for the supplied pubkey is in the state's pending_deposits queue. The
// returned witnesses are concatenated [deposit -> state_root,
// state_root -> block_header_root], matching ValidatorAndSlotProof.
func (state *BeaconState) PendingDepositProof(pubkey []byte) (witnesses [][]byte, depositIndex uint64, deposit *generic.PendingDeposit, err error) {

	// FIFO: pick the lowest-indexed match, which is the next entry the chain
	// will process.
	foundIndex := -1
	for i, pd := range state.PendingDeposits {
		if pd == nil {
			continue
		}
		if bytes.Equal(pd.Pubkey, pubkey) {
			foundIndex = i
			deposit = pd
			break
		}
	}
	if foundIndex < 0 {
		return nil, 0, nil, fmt.Errorf("no pending deposit found for pubkey 0x%x", pubkey)
	}
	depositIndex = uint64(foundIndex)

	stateTree, err := state.GetTree()
	if err != nil {
		return nil, 0, nil, fmt.Errorf("could not get state tree: %w", err)
	}

	// Generalized index for pending_deposits[depositIndex]:
	//   state root                       (gid = 1)
	//   -> pending_deposits field        (gid = 1 * 64 + 34)
	//   -> list data root (mixin parent) (gid * 2)
	//   -> deposit at depositIndex       (gid * maxLength + depositIndex)
	gid := uint64(1)*beaconStateChunkCeil + generic.BeaconStatePendingDepositsFieldIndex
	gid = gid*2*generic.BeaconStatePendingDepositsMaxLength + depositIndex

	depositStateProof, err := stateTree.Prove(int(gid))
	if err != nil {
		return nil, 0, nil, fmt.Errorf("could not get proof for pending deposit at index %d: %w", depositIndex, err)
	}

	// Sanity check that the proof leaf matches the deposit's hash tree root.
	depositRoot, err := deposit.HashTreeRoot()
	if err != nil {
		return nil, 0, nil, fmt.Errorf("could not compute hash tree root for pending deposit: %w", err)
	}
	if !bytes.Equal(depositStateProof.Leaf, depositRoot[:]) {
		return nil, 0, nil, errors.New("proof leaf does not match expected pending deposit")
	}

	// Drop the state tree before building the block header proof so the GC
	// can reclaim the (substantial) state tree memory.
	stateTree = nil

	blockHeaderTree, err := state.LatestBlockHeader.GetTree()
	if err != nil {
		return nil, 0, nil, fmt.Errorf("could not get block header tree: %w", err)
	}
	blockHeaderProof, err := blockHeaderTree.Prove(int(generic.BeaconBlockHeaderStateRootGeneralizedIndex))
	if err != nil {
		return nil, 0, nil, fmt.Errorf("could not get proof for block header: %w", err)
	}

	witnesses = make([][]byte, 0, len(depositStateProof.Hashes)+len(blockHeaderProof.Hashes))
	witnesses = append(witnesses, depositStateProof.Hashes...)
	witnesses = append(witnesses, blockHeaderProof.Hashes...)

	return witnesses, depositIndex, deposit, nil
}
