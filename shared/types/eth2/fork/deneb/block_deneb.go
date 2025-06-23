package deneb

import "github.com/rocket-pool/smartnode/shared/types/eth2/generic"

// Important indices for proof generation:
const BeaconBlockBodyChunksCeil uint64 = 16

func (b *SignedBeaconBlock) ProveWithdrawal(indexInWithdrawalsArray uint64) ([][]byte, error) {
	tree, err := b.Block.GetTree()
	if err != nil {
		return nil, err
	}

	gid := uint64(1)
	// Navigate to the body
	gid = gid*generic.BeaconBlockChunksCeil + generic.BeaconBlockBodyIndex
	// Then to the ExecutionPayload
	gid = gid*BeaconBlockBodyChunksCeil + generic.BeaconBlockBodyExecutionPayloadIndex
	// Then to the withdrawals array
	gid = gid*generic.BeaconBlockBodyExecutionPayloadChunksCeil + generic.BeaconBlockBodyExecutionPayloadWithdrawalsIndex
	// Then to the array contents
	gid = gid * 2
	// Finally to the withdrawal in question
	gid = gid*generic.BeaconBlockWithdrawalsArrayMax + indexInWithdrawalsArray

	proof, err := tree.Prove(int(gid))
	if err != nil {
		return nil, err
	}

	return proof.Hashes, nil
}

// Types needed for withdrawal proofs
type BeaconBlock struct {
	Slot          uint64           `json:"slot"`
	ProposerIndex uint64           `json:"proposer_index"`
	ParentRoot    [32]byte         `json:"parent_root" ssz-size:"32"`
	StateRoot     [32]byte         `json:"state_root" ssz-size:"32"`
	Body          *BeaconBlockBody `json:"body"`
}

type SignedBeaconBlock struct {
	Block     *BeaconBlock `json:"message"`
	Signature []byte       `json:"signature" ssz-size:"96"`
}

type BeaconBlockBody struct {
	RandaoReveal          []byte                                `json:"randao_reveal" ssz-size:"96"`
	Eth1Data              *generic.Eth1Data                     `json:"eth1_data"`
	Graffiti              [32]byte                              `json:"graffiti" ssz-size:"32"`
	ProposerSlashings     []*generic.ProposerSlashing           `json:"proposer_slashings" ssz-max:"16"`
	AttesterSlashings     []*generic.AttesterSlashing           `json:"attester_slashings" ssz-max:"2"`
	Attestations          []*generic.Attestation                `json:"attestations" ssz-max:"128"`
	Deposits              []*generic.Deposit                    `json:"deposits" ssz-max:"16"`
	VoluntaryExits        []*generic.SignedVoluntaryExit        `json:"voluntary_exits" ssz-max:"16"`
	SyncAggregate         *generic.SyncAggregate                `json:"sync_aggregate"`
	ExecutionPayload      *ExecutionPayload                     `json:"execution_payload"`
	BlsToExecutionChanges []*generic.SignedBLSToExecutionChange `json:"bls_to_execution_changes" ssz-max:"16"`
	BlobKzgCommitments    [][48]byte                            `json:"blob_kzg_commitments" ssz-max:"4096,48" ssz-size:"?,48"`
}

type ExecutionPayload struct {
	ParentHash    [32]byte              `ssz-size:"32" json:"parent_hash"`
	FeeRecipient  [20]byte              `ssz-size:"20" json:"fee_recipient"`
	StateRoot     [32]byte              `ssz-size:"32" json:"state_root"`
	ReceiptsRoot  [32]byte              `ssz-size:"32" json:"receipts_root"`
	LogsBloom     [256]byte             `ssz-size:"256" json:"logs_bloom"`
	PrevRandao    [32]byte              `ssz-size:"32" json:"prev_randao"`
	BlockNumber   uint64                `json:"block_number"`
	GasLimit      uint64                `json:"gas_limit"`
	GasUsed       uint64                `json:"gas_used"`
	Timestamp     uint64                `json:"timestamp"`
	ExtraData     []byte                `ssz-max:"32" json:"extra_data"`
	BaseFeePerGas generic.Uint256       `ssz-size:"32" json:"base_fee_per_gas"`
	BlockHash     [32]byte              `ssz-size:"32" json:"block_hash"`
	Transactions  [][]byte              `ssz-max:"1048576,1073741824" ssz-size:"?,?" json:"transactions"`
	Withdrawals   []*generic.Withdrawal `json:"withdrawals" ssz-max:"16"`
	BlobGasUsed   uint64                `json:"blob_gas_used"`
	ExcessBlobGas uint64                `json:"excess_blob_gas"`
}

func (b *SignedBeaconBlock) HasExecutionPayload() bool {
	return b.Block.Body.ExecutionPayload != nil
}

func (b *SignedBeaconBlock) Withdrawals() []*generic.Withdrawal {
	return b.Block.Body.ExecutionPayload.Withdrawals
}
