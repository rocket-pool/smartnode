package gloas

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"sync/atomic"

	"github.com/rocket-pool/smartnode/shared/types/eth2/generic"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

// The Gloas BeaconState has 46 fields, so the next power of two is 64.
const beaconStateChunkCeil uint64 = 64

// New in Gloas (EIP-7732)
type Builder struct {
	Pubkey            []byte   `json:"pubkey" ssz-size:"48"`
	Version           uint8    `json:"version"`
	ExecutionAddress  [20]byte `json:"execution_address" ssz-size:"20"`
	Balance           uint64   `json:"balance"`
	DepositEpoch      uint64   `json:"deposit_epoch"`
	WithdrawableEpoch uint64   `json:"withdrawable_epoch"`
}

// New in Gloas (EIP-7732)
type BuilderPendingWithdrawal struct {
	FeeRecipient [20]byte `json:"fee_recipient" ssz-size:"20"`
	Amount       uint64   `json:"amount"`
	BuilderIndex uint64   `json:"builder_index"`
}

// New in Gloas (EIP-7732)
type BuilderPendingPayment struct {
	Weight        uint64                    `json:"weight"`
	Withdrawal    *BuilderPendingWithdrawal `json:"withdrawal"`
	ProposerIndex uint64                    `json:"proposer_index"`
}

// Adapted from the Fulu BeaconState to the Gloas spec
// (https://github.com/ethereum/consensus-specs/blob/master/specs/gloas/beacon-chain.md#beaconstate):
// - Removed LatestExecutionPayloadHeader, replaced (in place) by LatestBlockHash (EIP-7732)
// - Added the builder/ePBS fields after ProposerLookahead (EIP-7732)
type BeaconState struct {
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
	LatestBlockHash              [32]byte                     `json:"latest_block_hash" ssz-size:"32"` // New in Gloas (EIP-7732), replaces LatestExecutionPayloadHeader
	NextWithdrawalIndex          uint64                       `json:"next_withdrawal_index"`
	NextWithdrawalValidatorIndex uint64                       `json:"next_withdrawal_validator_index"`
	HistoricalSummaries          []*generic.HistoricalSummary `json:"historical_summaries" ssz-max:"16777216"`

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

	// New in Gloas (EIP-7732)
	Builders                     []*Builder `json:"builders,omitempty" ssz-max:"1099511627776"`
	NextWithdrawalBuilderIndex   uint64     `json:"next_withdrawal_builder_index"`
	ExecutionPayloadAvailability []byte     `json:"execution_payload_availability" ssz-size:"1024"` // Bitvector[SLOTS_PER_HISTORICAL_ROOT] (8192 bits)
	// Vector[BuilderPendingPayment, 2 * SLOTS_PER_EPOCH]. fastssz cannot generate code for
	// vectors of containers, so each element is kept as its 52-byte SSZ encoding, which
	// serializes identically. Decode an entry with BuilderPendingPayment.UnmarshalSSZ.
	BuilderPendingPayments     [][]byte                    `json:"builder_pending_payments" ssz-size:"64,52"`
	BuilderPendingWithdrawals  []*BuilderPendingWithdrawal `json:"builder_pending_withdrawals,omitempty" ssz-max:"134217728"`
	LatestExecutionPayloadBid  *ExecutionPayloadBid        `json:"latest_execution_payload_bid"`
	PayloadExpectedWithdrawals []*generic.Withdrawal       `json:"payload_expected_withdrawals,omitempty" ssz-max:"16"`
	// Vector[Vector[ValidatorIndex, PTC_SIZE], (2 + MIN_SEED_LOOKAHEAD) * SLOTS_PER_EPOCH],
	// i.e. 96 slots x 512 validator indices. fastssz cannot generate code for 2D uint64
	// vectors, so it is flattened to 96*512 = 49152 indices, which serializes identically.
	PtcWindow []uint64 `json:"ptc_window" ssz-size:"49152"`
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
	// There's 46 fields, so rounding up to the next power of two is 64, a left-aligned node
	// BeaconStateValidatorsIndex is the 11th field, so its generalized index is 64 + 11 = 75
	return math.GetPowerOfTwoCeil(getStateChunkSize()) + generic.BeaconStateValidatorsIndex
}

func GetGeneralizedIndexForSlot() uint64 {
	// There's 46 fields, so rounding up to the next power of two is 64, a left-aligned node
	// BeaconStateSlotIndex is the 2nd field, so its generalized index is 64 + 2 = 66
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

func (state *BeaconState) GetPreviousEpochParticipation() []byte {
	return state.PreviousEpochParticipation
}
