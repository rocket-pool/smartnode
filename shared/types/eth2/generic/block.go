package generic

const BeaconBlockChunksCeil uint64 = 8
const BeaconBlockBodyIndex uint64 = 4
const BeaconBlockBodyExecutionPayloadIndex uint64 = 9
const BeaconBlockBodyExecutionPayloadChunksCeil uint64 = 32
const BeaconBlockBodyExecutionPayloadWithdrawalsIndex uint64 = 14
const BeaconBlockWithdrawalsArrayMax uint64 = 16

type ProposerSlashing struct {
	Header1 *SignedBeaconBlockHeader `json:"signed_header_1"`
	Header2 *SignedBeaconBlockHeader `json:"signed_header_2"`
}

type SignedBeaconBlockHeader struct {
	Header    *BeaconBlockHeader `json:"message"`
	Signature []byte             `json:"signature" ssz-size:"96"`
}

type AttesterSlashing struct {
	Attestation1 *IndexedAttestation `json:"attestation_1"`
	Attestation2 *IndexedAttestation `json:"attestation_2"`
}

type IndexedAttestation struct {
	AttestationIndices []uint64         `json:"attesting_indices" ssz-max:"2048"`
	Data               *AttestationData `json:"data"`
	Signature          []byte           `json:"signature" ssz-size:"96"`
}

type AttestationData struct {
	Slot            uint64      `json:"slot"`
	Index           uint64      `json:"index"`
	BeaconBlockHash [32]byte    `json:"beacon_block_root" ssz-size:"32"`
	Source          *Checkpoint `json:"source"`
	Target          *Checkpoint `json:"target"`
}

type Attestation struct {
	AggregationBits []byte           `json:"aggregation_bits" ssz:"bitlist" ssz-max:"2048"`
	Data            *AttestationData `json:"data"`
	Signature       [96]byte         `json:"signature" ssz-size:"96"`
}

type Deposit struct {
	Proof [][]byte `ssz-size:"33,32"`
	Data  *DepositData
}

type SignedVoluntaryExit struct {
	Exit      *VoluntaryExit `json:"message"`
	Signature [96]byte       `json:"signature" ssz-size:"96"`
}

type SyncAggregate struct {
	SyncCommiteeBits      []byte   `json:"sync_committee_bits" ssz-size:"64"`
	SyncCommiteeSignature [96]byte `json:"sync_committee_signature" ssz-size:"96"`
}

type ExecutionPayload struct {
	ParentHash    [32]byte      `ssz-size:"32" json:"parent_hash"`
	FeeRecipient  [20]byte      `ssz-size:"20" json:"fee_recipient"`
	StateRoot     [32]byte      `ssz-size:"32" json:"state_root"`
	ReceiptsRoot  [32]byte      `ssz-size:"32" json:"receipts_root"`
	LogsBloom     [256]byte     `ssz-size:"256" json:"logs_bloom"`
	PrevRandao    [32]byte      `ssz-size:"32" json:"prev_randao"`
	BlockNumber   uint64        `json:"block_number"`
	GasLimit      uint64        `json:"gas_limit"`
	GasUsed       uint64        `json:"gas_used"`
	Timestamp     uint64        `json:"timestamp"`
	ExtraData     []byte        `ssz-max:"32" json:"extra_data"`
	BaseFeePerGas Uint256       `ssz-size:"32" json:"base_fee_per_gas"`
	BlockHash     [32]byte      `ssz-size:"32" json:"block_hash"`
	Transactions  [][]byte      `ssz-max:"1048576,1073741824" ssz-size:"?,?" json:"transactions"`
	Withdrawals   []*Withdrawal `json:"withdrawals" ssz-max:"16"`
	BlobGasUsed   uint64        `json:"blob_gas_used"`
	ExcessBlobGas uint64        `json:"excess_blob_gas"`
}

type Withdrawal struct {
	Index          uint64   `json:"index"`
	ValidatorIndex uint64   `json:"validator_index"`
	Address        [20]byte `json:"address" ssz-size:"20"`
	Amount         uint64   `json:"amount"`
}

type BLSToExecutionChange struct {
	ValidatorIndex     uint64   `json:"validator_index"`
	FromBLSPubKey      [48]byte `json:"from_bls_pubkey" ssz-size:"48"`
	ToExecutionAddress [20]byte `json:"to_execution_address" ssz-size:"20"`
}

type SignedBLSToExecutionChange struct {
	Message   *BLSToExecutionChange `json:"message"`
	Signature [96]byte              `json:"signature" ssz-size:"96"`
}
