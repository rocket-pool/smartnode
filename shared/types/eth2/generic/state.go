package generic

// BeaconStateValidatorsIndex is the field offset of the Validators field in the BeaconState struct
const BeaconStateValidatorsIndex uint64 = 11

// SlotIndex is the field offset of the Slot field in the BeaconState struct
const BeaconStateSlotIndex uint64 = 2

// If this ever isn't a power of two, we need to round up to the next power of two
const beaconStateValidatorsMaxLength uint64 = 1 << 40

const BeaconStateHistoricalSummariesFieldIndex uint64 = 27
const BeaconStateHistoricalSummariesMaxLength uint64 = 1 << 24
const beaconStateHistoricalSummaryChunkCeil uint64 = 2
const beaconStateHistoricalSummaryBlockSummaryRootIndex uint64 = 0
const BeaconStateBlockRootsMaxLength uint64 = 1 << 13
const BeaconStateBlockRootsFieldIndex uint64 = 5

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
