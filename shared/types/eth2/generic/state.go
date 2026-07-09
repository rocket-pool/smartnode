package generic

// BeaconStateValidatorsIndex is the field offset of the Validators field in the BeaconState struct
const BeaconStateValidatorsIndex uint64 = 11

// SlotIndex is the field offset of the Slot field in the BeaconState struct
const BeaconStateSlotIndex uint64 = 2

// If this ever isn't a power of two, we need to round up to the next power of two
const beaconStateValidatorsMaxLength uint64 = 1 << 40

const BeaconStateHistoricalSummariesFieldIndex uint64 = 27
const BeaconStateHistoricalSummariesMaxLength uint64 = 1 << 24
const BeaconStateBlockRootsMaxLength uint64 = 1 << 13
const BeaconStateBlockRootsFieldIndex uint64 = 5

// BeaconStatePreviousEpochParticipationFieldIndex is the field offset of the
// PreviousEpochParticipation field in the BeaconState struct
const BeaconStatePreviousEpochParticipationFieldIndex uint64 = 15

// BeaconStateParticipationMaxChunks is the chunk count of the epoch
// participation byte lists: VALIDATOR_REGISTRY_LIMIT (2^40) one-byte flags
// packed 32 per chunk.
const BeaconStateParticipationMaxChunks uint64 = 1 << 35

// GetGeneralizedIndexForParticipationChunk returns the generalized index of
// the 32-byte chunk of a participation byte list, starting from the
// generalized index of the list field itself.
func GetGeneralizedIndexForParticipationChunk(chunkIndex uint64, participationFieldGid uint64) uint64 {
	// Lists have a base index of 2 (the length mixin occupies the sibling).
	return participationFieldGid*2*BeaconStateParticipationMaxChunks + chunkIndex
}

type PendingDeposit struct {
	Pubkey                []byte `json:"pubkey" ssz-size:"48"`
	WithdrawalCredentials []byte `json:"withdrawal_credentials" ssz-size:"32"`
	Amount                uint64 `json:"amount"`
	Signature             []byte `json:"signature" ssz-size:"96"`
	Slot                  uint64 `json:"slot"`
}

type PendingPartialWithdrawal struct {
	ValidatorIndex    uint64 `json:"validator_index"`
	Amount            uint64 `json:"amount"`
	WithdrawableEpoch uint64 `json:"withdrawable_epoch"`
}

type PendingConsolidation struct {
	SourceIndex uint64 `json:"source_index"`
	TargetIndex uint64 `json:"target_index"`
}

type ExecutionPayloadHeader struct {
	ParentHash       [32]byte  `json:"parent_hash" ssz-size:"32"`
	FeeRecipient     [20]byte  `json:"fee_recipient" ssz-size:"20"`
	StateRoot        [32]byte  `json:"state_root" ssz-size:"32"`
	ReceiptsRoot     [32]byte  `json:"receipts_root" ssz-size:"32"`
	LogsBloom        [256]byte `json:"logs_bloom" ssz-size:"256"`
	PrevRandao       [32]byte  `json:"prev_randao" ssz-size:"32"`
	BlockNumber      uint64    `json:"block_number"`
	GasLimit         uint64    `json:"gas_limit"`
	GasUsed          uint64    `json:"gas_used"`
	Timestamp        uint64    `json:"timestamp"`
	ExtraData        []byte    `json:"extra_data" ssz-max:"32"`
	BaseFeePerGas    Uint256   `json:"base_fee_per_gas" ssz-size:"32"`
	BlockHash        [32]byte  `json:"block_hash" ssz-size:"32"`
	TransactionsRoot [32]byte  `json:"transactions_root" ssz-size:"32"`
	WithdrawalRoot   [32]byte  `json:"withdrawals_root" ssz-size:"32"`
	BlobGasUsed      uint64    `json:"blob_gas_used"`
	ExcessBlobGas    uint64    `json:"excess_blob_gas"`
}
