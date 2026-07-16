package gloas

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/rocket-pool/smartnode/shared/types/eth2/generic"
)

// Gloas BeaconState field ssz-indices (EIP-7688 ProgressiveContainer).
// Must stay aligned with the BeaconState struct tags.
const (
	beaconStateSlotFieldIndex                = generic.BeaconStateSlotIndex                     // 2
	beaconStateBlockRootsFieldIndex          = generic.BeaconStateBlockRootsFieldIndex          // 5
	beaconStateValidatorsFieldIndex          = generic.BeaconStateValidatorsIndex               // 11
	beaconStateHistoricalSummariesFieldIndex = generic.BeaconStateHistoricalSummariesFieldIndex // 27
)

// New in Gloas (EIP-7732). Immutable container (not ProgressiveContainer).
type Builder struct {
	Pubkey            []byte   `json:"pubkey" ssz-size:"48"`
	Version           uint8    `json:"version"`
	ExecutionAddress  [20]byte `json:"execution_address" ssz-size:"20"`
	Balance           uint64   `json:"balance"`
	DepositEpoch      uint64   `json:"deposit_epoch"`
	WithdrawableEpoch uint64   `json:"withdrawable_epoch"`
}

// New in Gloas (EIP-7732). Immutable container (not ProgressiveContainer).
type BuilderPendingWithdrawal struct {
	FeeRecipient [20]byte `json:"fee_recipient" ssz-size:"20"`
	Amount       uint64   `json:"amount"`
	BuilderIndex uint64   `json:"builder_index"`
}

// New in Gloas (EIP-7732). Immutable container (not ProgressiveContainer).
type BuilderPendingPayment struct {
	Weight        uint64                    `json:"weight"`
	Withdrawal    *BuilderPendingWithdrawal `json:"withdrawal"`
	ProposerIndex uint64                    `json:"proposer_index"`
}

// Adapted from the Fulu BeaconState to the Gloas spec
// (https://github.com/ethereum/consensus-specs/blob/master/specs/gloas/beacon-chain.md#beaconstate):
//   - Removed LatestExecutionPayloadHeader, replaced (in place) by LatestBlockHash (EIP-7732)
//   - Added the builder/ePBS fields after ProposerLookahead (EIP-7732)
//   - EIP-7688: ProgressiveContainer + ProgressiveList on evolving fields
type BeaconState struct {
	GenesisTime                  uint64                       `json:"genesis_time" ssz-index:"0"`
	GenesisValidatorsRoot        []byte                       `json:"genesis_validators_root" ssz-size:"32" ssz-index:"1"`
	Slot                         uint64                       `json:"slot" ssz-index:"2"`
	Fork                         *generic.Fork                `json:"fork" ssz-index:"3"`
	LatestBlockHeader            *generic.BeaconBlockHeader   `json:"latest_block_header" ssz-index:"4"`
	BlockRoots                   [8192][32]byte               `json:"block_roots" ssz-size:"8192,32" ssz-index:"5"`
	StateRoots                   [8192][32]byte               `json:"state_roots" ssz-size:"8192,32" ssz-index:"6"`
	HistoricalRoots              [][]byte                     `json:"historical_roots" ssz-max:"16777216" ssz-size:"?,32" ssz-index:"7"`
	Eth1Data                     *generic.Eth1Data            `json:"eth1_data" ssz-index:"8"`
	Eth1DataVotes                []*generic.Eth1Data          `json:"eth1_data_votes" ssz-max:"2048" ssz-index:"9"`
	Eth1DepositIndex             uint64                       `json:"eth1_deposit_index" ssz-index:"10"`
	Validators                   []*generic.Validator         `json:"validators" ssz-type:"progressive-list" ssz-max:"1099511627776" ssz-index:"11"`
	Balances                     []uint64                     `json:"balances" ssz-type:"progressive-list" ssz-max:"1099511627776" ssz-index:"12"`
	RandaoMixes                  [][]byte                     `json:"randao_mixes" ssz-size:"65536,32" ssz-index:"13"`
	Slashings                    []uint64                     `json:"slashings" ssz-size:"8192" ssz-index:"14"`
	PreviousEpochParticipation   []byte                       `json:"previous_epoch_participation" ssz-type:"progressive-list" ssz-max:"1099511627776" ssz-index:"15"`
	CurrentEpochParticipation    []byte                       `json:"current_epoch_participation" ssz-type:"progressive-list" ssz-max:"1099511627776" ssz-index:"16"`
	JustificationBits            [1]byte                      `json:"justification_bits" ssz-size:"1" ssz-index:"17"`
	PreviousJustifiedCheckpoint  *generic.Checkpoint          `json:"previous_justified_checkpoint" ssz-index:"18"`
	CurrentJustifiedCheckpoint   *generic.Checkpoint          `json:"current_justified_checkpoint" ssz-index:"19"`
	FinalizedCheckpoint          *generic.Checkpoint          `json:"finalized_checkpoint" ssz-index:"20"`
	InactivityScores             []uint64                     `json:"inactivity_scores" ssz-type:"progressive-list" ssz-max:"1099511627776" ssz-index:"21"`
	CurrentSyncCommittee         *generic.SyncCommittee       `json:"current_sync_committee" ssz-index:"22"`
	NextSyncCommittee            *generic.SyncCommittee       `json:"next_sync_committee" ssz-index:"23"`
	LatestBlockHash              [32]byte                     `json:"latest_block_hash" ssz-size:"32" ssz-index:"24"` // New in Gloas (EIP-7732)
	NextWithdrawalIndex          uint64                       `json:"next_withdrawal_index" ssz-index:"25"`
	NextWithdrawalValidatorIndex uint64                       `json:"next_withdrawal_validator_index" ssz-index:"26"`
	HistoricalSummaries          []*generic.HistoricalSummary `json:"historical_summaries" ssz-max:"16777216" ssz-index:"27"`

	// New in Electra
	DepositRequestsStartIndex     uint64                              `json:"deposit_requests_start_index" ssz-index:"28"`
	DepositBalanceToConsume       uint64                              `json:"deposit_balance_to_consume" ssz-index:"29"`
	ExitBalanceToConsume          uint64                              `json:"exit_balance_to_consume" ssz-index:"30"`
	EarliestExitEpoch             uint64                              `json:"earliest_exit_epoch" ssz-index:"31"`
	ConsolidationBalanceToConsume uint64                              `json:"consolidation_balance_to_consume" ssz-index:"32"`
	EarliestConsolidationEpoch    uint64                              `json:"earliest_consolidation_epoch" ssz-index:"33"`
	PendingDeposits               []*generic.PendingDeposit           `json:"pending_deposits,omitempty" ssz-type:"progressive-list" ssz-max:"134217728" ssz-index:"34"`
	PendingPartialWithdrawals     []*generic.PendingPartialWithdrawal `json:"pending_partial_withdrawals,omitempty" ssz-type:"progressive-list" ssz-max:"134217728" ssz-index:"35"`
	PendingConsolidations         []*generic.PendingConsolidation     `json:"pending_consolidations,omitempty" ssz-type:"progressive-list" ssz-max:"262144" ssz-index:"36"`

	// New in Fulu
	ProposerLookahead []uint64 `json:"proposer_lookahead,omitempty" ssz-size:"64" ssz-index:"37"`

	// New in Gloas (EIP-7732)
	Builders                     []*Builder                  `json:"builders,omitempty" ssz-type:"progressive-list" ssz-max:"1099511627776" ssz-index:"38"`
	NextWithdrawalBuilderIndex   uint64                      `json:"next_withdrawal_builder_index" ssz-index:"39"`
	ExecutionPayloadAvailability []byte                      `json:"execution_payload_availability" ssz-size:"1024" ssz-index:"40"` // Bitvector[SLOTS_PER_HISTORICAL_ROOT]
	BuilderPendingPayments       [64]*BuilderPendingPayment  `json:"builder_pending_payments" ssz-size:"64" ssz-index:"41"`         // Vector[BuilderPendingPayment, 2 * SLOTS_PER_EPOCH]
	BuilderPendingWithdrawals    []*BuilderPendingWithdrawal `json:"builder_pending_withdrawals,omitempty" ssz-type:"progressive-list" ssz-max:"134217728" ssz-index:"42"`
	LatestExecutionPayloadBid    *ExecutionPayloadBid        `json:"latest_execution_payload_bid" ssz-index:"43"`
	PayloadExpectedWithdrawals   []*generic.Withdrawal       `json:"payload_expected_withdrawals,omitempty" ssz-type:"progressive-list" ssz-max:"16" ssz-index:"44"`
	// Vector[Vector[ValidatorIndex, PTC_SIZE], (2 + MIN_SEED_LOOKAHEAD) * SLOTS_PER_EPOCH]
	// = 96 slots x 512 validator indices.
	PtcWindow [96][512]uint64 `json:"ptc_window" ssz-size:"96,512" ssz-index:"45"`
}

// GetGeneralizedIndexForValidators returns the gindex of the validators field
// root inside the Gloas ProgressiveContainer BeaconState.
func GetGeneralizedIndexForValidators() uint64 {
	return generic.ProgressiveContainerFieldGindex(beaconStateValidatorsFieldIndex)
}

// GetGeneralizedIndexForSlot returns the gindex of the slot field inside the
// Gloas ProgressiveContainer BeaconState.
func GetGeneralizedIndexForSlot() uint64 {
	return generic.ProgressiveContainerFieldGindex(beaconStateSlotFieldIndex)
}

// GetGeneralizedIndexForBlockRoots returns the gindex of the block_roots field
// root inside the Gloas ProgressiveContainer BeaconState.
func GetGeneralizedIndexForBlockRoots() uint64 {
	return generic.ProgressiveContainerFieldGindex(beaconStateBlockRootsFieldIndex)
}

// GetGeneralizedIndexForHistoricalSummaries returns the gindex of the
// historical_summaries field root inside the Gloas ProgressiveContainer BeaconState.
// historical_summaries remains a normal List (not ProgressiveList) per EIP-7688.
func GetGeneralizedIndexForHistoricalSummaries() uint64 {
	return generic.ProgressiveContainerFieldGindex(beaconStateHistoricalSummariesFieldIndex)
}

// GetGeneralizedIndexForValidator returns the gindex of validators[index] in a
// Gloas BeaconState (ProgressiveContainer field + ProgressiveList element).
//
// Do not use generic.GetGeneralizedIndexForValidator here — that is the classic
// fixed-capacity List formula for pre-Gloas states (Deneb/Electra/Fulu).
func GetGeneralizedIndexForValidator(validatorIndex uint64) uint64 {
	return generic.GetGeneralizedIndexForProgressiveListElement(
		GetGeneralizedIndexForValidators(),
		validatorIndex,
	)
}

// ValidatorAndSlotProof produces both the validator proof and the slot proof
// for the state's current slot, using EIP-7688 progressive g-indices.
func (state *BeaconState) ValidatorAndSlotProof(validatorIndex uint64) ([][]byte, [][]byte, error) {

	if validatorIndex >= uint64(len(state.Validators)) {
		return nil, nil, errors.New("validator index out of bounds")
	}

	stateTree, err := generic.SSZ.GetTree(state)
	if err != nil {
		return nil, nil, fmt.Errorf("could not get state tree: %w", err)
	}

	validatorGid := GetGeneralizedIndexForValidator(validatorIndex)
	validatorStateProof, err := stateTree.Prove(int(validatorGid))
	if err != nil {
		return nil, nil, fmt.Errorf("could not get proof for validator: %w", err)
	}

	// Sanity check that the proof leaf matches the expected validator
	validatorHashTreeRoot, err := generic.SSZ.HashTreeRoot(state.Validators[validatorIndex])
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

	bhTree, err := generic.SSZ.GetTree(state.LatestBlockHeader)
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
	root, err := generic.SSZ.GetTree(blockHeader)
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
	tree, err := generic.SSZ.GetTree(state)
	if err != nil {
		return nil, fmt.Errorf("could not get state tree: %w", err)
	}

	// ProgressiveContainer field → historical_summaries (still a fixed-capacity List).
	arrayIndex := (slot / generic.SlotsPerHistoricalRoot) - capellaOffset
	gid := generic.GetGeneralizedIndexForListElement(
		GetGeneralizedIndexForHistoricalSummaries(),
		generic.BeaconStateHistoricalSummariesMaxLength,
		arrayIndex,
	)

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
	tree, err := generic.SSZ.GetTree(&hsls)
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

	tree, err := generic.SSZ.GetTree(state)
	if err != nil {
		return nil, fmt.Errorf("could not get state tree: %w", err)
	}

	// ProgressiveContainer field → block_roots Vector[Root, SLOTS_PER_HISTORICAL_ROOT].
	gid := generic.GetGeneralizedIndexForVectorElement(
		GetGeneralizedIndexForBlockRoots(),
		generic.BeaconStateBlockRootsMaxLength,
		slot%generic.SlotsPerHistoricalRoot,
	)

	proof, err := tree.Prove(int(gid))
	if err != nil {
		return nil, fmt.Errorf("could not get proof for block root: %w", err)
	}

	return proof.Hashes, nil
}

func (state *BeaconState) BlockHeaderProof() ([][]byte, error) {
	// Construct block header with state root
	stateRoot, err := generic.SSZ.HashTreeRoot(state)
	if err != nil {
		return nil, fmt.Errorf("could not get state root: %w", err)
	}
	latestBlockHeader := state.LatestBlockHeader
	blockHeader := *latestBlockHeader
	blockHeader.StateRoot = stateRoot[:]
	blockHeaderTree, err := generic.SSZ.GetTree(&blockHeader)
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
