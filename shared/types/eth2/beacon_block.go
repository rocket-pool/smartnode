package eth2

// Important indices for proof generation:
// beaconBlockDenebBodyIndex is the field offset of the Body field in the BeaconBlockDeneb struct
const beaconBlockDenebChunksCeil int = 8
const beaconBlockDenebBodyIndex int = 4
const beaconBlockDenebBodyChunksCeil int = 16
const beaconBlockDenebBodyExecutionPayloadIndex int = 9
const beaconBlockDenebBodyExecutionPayloadChunksCeil int = 32
const beaconBlockDenebBodyExecutionPayloadWithdrawalsIndex int = 14
const beaconBlockDenebWithdrawalsArrayMax int = 16

func (b *BeaconBlockDeneb) ProveWithdrawal(indexInWithdrawalsArray int) ([][]byte, error) {
	tree, err := b.GetTree()
	if err != nil {
		return nil, err
	}

	gid := 1
	// Navigate to the body
	gid = gid*beaconBlockDenebChunksCeil + beaconBlockDenebBodyIndex
	// Then to the ExecutionPayload
	gid = gid*beaconBlockDenebBodyChunksCeil + beaconBlockDenebBodyExecutionPayloadIndex
	// Then to the withdrawals array
	gid = gid*beaconBlockDenebBodyExecutionPayloadChunksCeil + beaconBlockDenebBodyExecutionPayloadWithdrawalsIndex
	// Then to the array contents
	gid = gid * 2
	// Finally to the withdrawal in question
	gid = gid*beaconBlockDenebWithdrawalsArrayMax + indexInWithdrawalsArray

	proof, err := tree.Prove(gid)
	if err != nil {
		return nil, err
	}

	return proof.Hashes, nil
}

// Types needed for withdrawal proofs
type BeaconBlockDeneb struct {
	Slot          uint64                `json:"slot"`
	ProposerIndex uint64                `json:"proposer_index"`
	ParentRoot    [32]byte              `json:"parent_root" ssz-size:"32"`
	StateRoot     [32]byte              `json:"state_root" ssz-size:"32"`
	Body          *BeaconBlockBodyDeneb `json:"body"`
}

type SignedBeaconBlockDeneb struct {
	Block     *BeaconBlockDeneb `json:"message"`
	Signature []byte            `json:"signature" ssz-size:"96"`
}

type BeaconBlockBodyDeneb struct {
	RandaoReveal          []byte                        `json:"randao_reveal" ssz-size:"96"`
	Eth1Data              *Eth1Data                     `json:"eth1_data"`
	Graffiti              [32]byte                      `json:"graffiti" ssz-size:"32"`
	ProposerSlashings     []*ProposerSlashing           `json:"proposer_slashings" ssz-max:"16"`
	AttesterSlashings     []*AttesterSlashing           `json:"attester_slashings" ssz-max:"2"`
	Attestations          []*Attestation                `json:"attestations" ssz-max:"128"`
	Deposits              []*Deposit                    `json:"deposits" ssz-max:"16"`
	VoluntaryExits        []*SignedVoluntaryExit        `json:"voluntary_exits" ssz-max:"16"`
	SyncAggregate         *SyncAggregate                `json:"sync_aggregate"`
	ExecutionPayload      *ExecutionPayloadDeneb        `json:"execution_payload"`
	BlsToExecutionChanges []*SignedBLSToExecutionChange `json:"bls_to_execution_changes" ssz-max:"16"`
	BlobKzgCommitments    [][48]byte                    `json:"blob_kzg_commitments" ssz-max:"4096,48" ssz-size:"?,48"`
}

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

type ExecutionPayloadDeneb struct {
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
