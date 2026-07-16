package gloas

import (
	"errors"

	"github.com/prysmaticlabs/go-bitfield"

	"github.com/rocket-pool/smartnode/shared/types/eth2/generic"
)

// Important indices for proof generation:
// The Gloas BeaconBlockBody has 13 fields. Pre-EIP-7688 proofs used power-of-two
// body layout (ceil 16). Under EIP-7688 BeaconBlockBody is a ProgressiveContainer.
const BeaconBlockBodyChunksCeil uint64 = 16

// In Gloas (EIP-7732), the execution payload is no longer part of the beacon block.
// It is distributed separately in a SignedExecutionPayloadEnvelope, and the block body
// only carries a SignedExecutionPayloadBid. Withdrawals can no longer be proven
// against the block root; they are committed to in the state's payload_expected_withdrawals.
func (b *SignedBeaconBlock) ProveWithdrawal(indexInWithdrawalsArray uint64) ([][]byte, error) {
	return nil, errors.New("gloas blocks do not contain the execution payload; withdrawals must be proven against the beacon state's payload_expected_withdrawals field")
}

// BeaconBlock remains a normal Container (not converted by EIP-7688).
type BeaconBlock struct {
	Slot          uint64           `json:"slot"`
	ProposerIndex uint64           `json:"proposer_index"`
	ParentRoot    [32]byte         `json:"parent_root" ssz-size:"32"`
	StateRoot     [32]byte         `json:"state_root" ssz-size:"32"`
	Body          *BeaconBlockBody `json:"body"`
}

// SignedBeaconBlock remains a normal Container (not converted by EIP-7688).
type SignedBeaconBlock struct {
	Block     *BeaconBlock `json:"message"`
	Signature []byte       `json:"signature" ssz-size:"96"`
}

// Modified in Gloas (EIP-7732/EIP-7688):
//   - Removed ExecutionPayload, BlobKzgCommitments and ExecutionRequests (moved to the ExecutionPayloadEnvelope)
//   - Added SignedExecutionPayloadBid, PayloadAttestations and ParentExecutionRequests
//   - EIP-7688: ProgressiveContainer + ProgressiveList on operation lists
type BeaconBlockBody struct {
	RandaoReveal              []byte                                `json:"randao_reveal" ssz-size:"96" ssz-index:"0"`
	Eth1Data                  *generic.Eth1Data                     `json:"eth1_data" ssz-index:"1"`
	Graffiti                  [32]byte                              `json:"graffiti" ssz-size:"32" ssz-index:"2"`
	ProposerSlashings         []*generic.ProposerSlashing           `json:"proposer_slashings" ssz-type:"progressive-list" ssz-max:"16" ssz-index:"3"`
	AttesterSlashings         []*AttesterSlashing                   `json:"attester_slashings" ssz-type:"progressive-list" ssz-max:"1" ssz-index:"4"`
	Attestations              []*Attestation                        `json:"attestations" ssz-type:"progressive-list" ssz-max:"8" ssz-index:"5"`
	Deposits                  []*generic.Deposit                    `json:"deposits" ssz-type:"progressive-list" ssz-max:"16" ssz-index:"6"`
	VoluntaryExits            []*generic.SignedVoluntaryExit        `json:"voluntary_exits" ssz-type:"progressive-list" ssz-max:"16" ssz-index:"7"`
	SyncAggregate             *generic.SyncAggregate                `json:"sync_aggregate" ssz-index:"8"`
	BlsToExecutionChanges     []*generic.SignedBLSToExecutionChange `json:"bls_to_execution_changes" ssz-type:"progressive-list" ssz-max:"16" ssz-index:"9"`
	SignedExecutionPayloadBid *SignedExecutionPayloadBid            `json:"signed_execution_payload_bid" ssz-index:"10"`
	PayloadAttestations       []*PayloadAttestation                 `json:"payload_attestations" ssz-type:"progressive-list" ssz-max:"4" ssz-index:"11"`
	ParentExecutionRequests   *ExecutionRequests                    `json:"parent_execution_requests" ssz-index:"12"`
}

// Modified in Gloas:EIP-7688 — ProgressiveContainer with ProgressiveBitlist aggregation_bits.
type Attestation struct {
	AggregationBits bitfield.Bitlist         `json:"aggregation_bits" ssz-type:"progressive-bitlist" ssz-max:"131072" ssz-index:"0"`
	Data            *generic.AttestationData `json:"data" ssz-index:"1"`
	Signature       [96]byte                 `json:"signature" ssz-size:"96" ssz-index:"2"`
	CommitteeBits   []byte                   `json:"committee_bits" ssz-size:"8" ssz-index:"3"`
}

// New in Gloas (EIP-7732). ProgressiveContainer (EIP-7688).
type ExecutionPayloadBid struct {
	ParentBlockHash       [32]byte   `json:"parent_block_hash" ssz-size:"32" ssz-index:"0"`
	ParentBlockRoot       [32]byte   `json:"parent_block_root" ssz-size:"32" ssz-index:"1"`
	BlockHash             [32]byte   `json:"block_hash" ssz-size:"32" ssz-index:"2"`
	PrevRandao            [32]byte   `json:"prev_randao" ssz-size:"32" ssz-index:"3"`
	FeeRecipient          [20]byte   `json:"fee_recipient" ssz-size:"20" ssz-index:"4"`
	GasLimit              uint64     `json:"gas_limit" ssz-index:"5"`
	BuilderIndex          uint64     `json:"builder_index" ssz-index:"6"`
	Slot                  uint64     `json:"slot" ssz-index:"7"`
	Value                 uint64     `json:"value" ssz-index:"8"`
	ExecutionPayment      uint64     `json:"execution_payment" ssz-index:"9"`
	BlobKzgCommitments    [][48]byte `json:"blob_kzg_commitments" ssz-type:"progressive-list" ssz-max:"4096" ssz-size:"?,48" ssz-index:"10"`
	ExecutionRequestsRoot [32]byte   `json:"execution_requests_root" ssz-size:"32" ssz-index:"11"`
}

// New in Gloas (EIP-7732). Immutable container.
type SignedExecutionPayloadBid struct {
	Message   *ExecutionPayloadBid `json:"message"`
	Signature [96]byte             `json:"signature" ssz-size:"96"`
}

// New in Gloas (EIP-7732). Immutable container.
type PayloadAttestationData struct {
	BeaconBlockRoot   [32]byte `json:"beacon_block_root" ssz-size:"32"`
	Slot              uint64   `json:"slot"`
	PayloadPresent    bool     `json:"payload_present"`
	BlobDataAvailable bool     `json:"blob_data_available"`
}

// New in Gloas (EIP-7732). ProgressiveContainer.
// AggregationBits is a Bitvector[PTC_SIZE] where PTC_SIZE = 512 (64 bytes).
type PayloadAttestation struct {
	AggregationBits []byte                  `json:"aggregation_bits" ssz-size:"64" ssz-index:"0"`
	Data            *PayloadAttestationData `json:"data" ssz-index:"1"`
	Signature       [96]byte                `json:"signature" ssz-size:"96" ssz-index:"2"`
}

// Modified in Gloas (EIP-8282/EIP-7688): ProgressiveContainer with ProgressiveList fields;
// added BuilderDeposits and BuilderExits.
type ExecutionRequests struct {
	Deposits        []*DepositRequest        `json:"deposits" ssz-type:"progressive-list" ssz-max:"8192" ssz-index:"0"`
	Withdrawals     []*WithdrawalRequest     `json:"withdrawals" ssz-type:"progressive-list" ssz-max:"16" ssz-index:"1"`
	Consolidations  []*ConsolidationRequest  `json:"consolidations" ssz-type:"progressive-list" ssz-max:"2" ssz-index:"2"`
	BuilderDeposits []*BuilderDepositRequest `json:"builder_deposits" ssz-type:"progressive-list" ssz-max:"64" ssz-index:"3"`
	BuilderExits    []*BuilderExitRequest    `json:"builder_exits" ssz-type:"progressive-list" ssz-max:"16" ssz-index:"4"`
}

// Immutable container.
type DepositRequest struct {
	Pubkey                []byte `json:"pubkey" ssz-size:"48"`
	WithdrawalCredentials []byte `json:"withdrawal_credentials" ssz-size:"32"`
	Amount                uint64 `json:"amount"`
	Signature             []byte `json:"signature" ssz-size:"96"`
	Index                 uint64 `json:"index"`
}

// Immutable container.
type WithdrawalRequest struct {
	SourceAddress   []byte `json:"source_address" ssz-size:"20"`
	ValidatorPubkey []byte `json:"validator_pubkey" ssz-size:"48"`
	Amount          uint64 `json:"amount"`
}

// Immutable container.
type ConsolidationRequest struct {
	SourceAddress []byte `json:"source_address" ssz-size:"20"`
	SourcePubkey  []byte `json:"source_pubkey" ssz-size:"48"`
	TargetPubkey  []byte `json:"target_pubkey" ssz-size:"48"`
}

// New in Gloas (EIP-8282). Immutable container.
type BuilderDepositRequest struct {
	Pubkey                []byte `json:"pubkey" ssz-size:"48"`
	WithdrawalCredentials []byte `json:"withdrawal_credentials" ssz-size:"32"`
	Amount                uint64 `json:"amount"`
	Signature             []byte `json:"signature" ssz-size:"96"`
}

// New in Gloas (EIP-8282). Immutable container.
type BuilderExitRequest struct {
	SourceAddress []byte `json:"source_address" ssz-size:"20"`
	Pubkey        []byte `json:"pubkey" ssz-size:"48"`
}

// Immutable container (AttesterSlashing itself is listed as immutable in EIP-7688).
type AttesterSlashing struct {
	Attestation1 *IndexedAttestation `json:"attestation_1"`
	Attestation2 *IndexedAttestation `json:"attestation_2"`
}

// Modified in Gloas:EIP-7688 — ProgressiveContainer with ProgressiveList attesting_indices.
type IndexedAttestation struct {
	AttestingIndices []uint64                 `json:"attesting_indices" ssz-type:"progressive-list" ssz-max:"131072" ssz-index:"0"`
	Data             *generic.AttestationData `json:"data" ssz-index:"1"`
	Signature        []byte                   `json:"signature" ssz-size:"96" ssz-index:"2"`
}

// In Gloas the execution payload is never part of the block itself.
func (b *SignedBeaconBlock) HasExecutionPayload() bool {
	return false
}

func (b *SignedBeaconBlock) Withdrawals() []*generic.Withdrawal {
	return nil
}
