package gloas

import (
	"errors"

	"github.com/prysmaticlabs/go-bitfield"

	"github.com/rocket-pool/smartnode/shared/types/eth2/generic"
)

// Important indices for proof generation:
// The Gloas BeaconBlockBody has 13 fields, so the next power of two is 16.
const BeaconBlockBodyChunksCeil uint64 = 16

// In Gloas (EIP-7732), the execution payload is no longer part of the beacon block.
// It is distributed separately in a SignedExecutionPayloadEnvelope, and the block body
// only carries a SignedExecutionPayloadBid. Withdrawals can no longer be proven
// against the block root; they are committed to in the state's payload_expected_withdrawals.
func (b *SignedBeaconBlock) ProveWithdrawal(indexInWithdrawalsArray uint64) ([][]byte, error) {
	return nil, errors.New("gloas blocks do not contain the execution payload; withdrawals must be proven against the beacon state's payload_expected_withdrawals field")
}

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

// Modified in Gloas (EIP-7732/EIP-7688):
// - Removed ExecutionPayload, BlobKzgCommitments and ExecutionRequests (moved to the ExecutionPayloadEnvelope)
// - Added SignedExecutionPayloadBid, PayloadAttestations and ParentExecutionRequests
type BeaconBlockBody struct {
	RandaoReveal              []byte                                `json:"randao_reveal" ssz-size:"96"`
	Eth1Data                  *generic.Eth1Data                     `json:"eth1_data"`
	Graffiti                  [32]byte                              `json:"graffiti" ssz-size:"32"`
	ProposerSlashings         []*generic.ProposerSlashing           `json:"proposer_slashings" ssz-max:"16"`
	AttesterSlashings         []*AttesterSlashing                   `json:"attester_slashings" ssz-max:"1"`
	Attestations              []*Attestation                        `json:"attestations" ssz-max:"8"`
	Deposits                  []*generic.Deposit                    `json:"deposits" ssz-max:"16"`
	VoluntaryExits            []*generic.SignedVoluntaryExit        `json:"voluntary_exits" ssz-max:"16"`
	SyncAggregate             *generic.SyncAggregate                `json:"sync_aggregate"`
	BlsToExecutionChanges     []*generic.SignedBLSToExecutionChange `json:"bls_to_execution_changes" ssz-max:"16"`
	SignedExecutionPayloadBid *SignedExecutionPayloadBid            `json:"signed_execution_payload_bid"`
	PayloadAttestations       []*PayloadAttestation                 `json:"payload_attestations" ssz-max:"4"`
	ParentExecutionRequests   *ExecutionRequests                    `json:"parent_execution_requests"`
}

type Attestation struct {
	AggregationBits bitfield.Bitlist         `json:"aggregation_bits" ssz:"bitlist" ssz-max:"131072"`
	Data            *generic.AttestationData `json:"data"`
	Signature       [96]byte                 `json:"signature" ssz-size:"96"`
	CommitteeBits   []byte                   `json:"committee_bits" ssz-size:"8"`
}

// New in Gloas (EIP-7732)
type ExecutionPayloadBid struct {
	ParentBlockHash       [32]byte   `json:"parent_block_hash" ssz-size:"32"`
	ParentBlockRoot       [32]byte   `json:"parent_block_root" ssz-size:"32"`
	BlockHash             [32]byte   `json:"block_hash" ssz-size:"32"`
	PrevRandao            [32]byte   `json:"prev_randao" ssz-size:"32"`
	FeeRecipient          [20]byte   `json:"fee_recipient" ssz-size:"20"`
	GasLimit              uint64     `json:"gas_limit"`
	BuilderIndex          uint64     `json:"builder_index"`
	Slot                  uint64     `json:"slot"`
	Value                 uint64     `json:"value"`
	ExecutionPayment      uint64     `json:"execution_payment"`
	BlobKzgCommitments    [][48]byte `json:"blob_kzg_commitments" ssz-max:"4096,48" ssz-size:"?,48"`
	ExecutionRequestsRoot [32]byte   `json:"execution_requests_root" ssz-size:"32"`
}

// New in Gloas (EIP-7732)
type SignedExecutionPayloadBid struct {
	Message   *ExecutionPayloadBid `json:"message"`
	Signature [96]byte             `json:"signature" ssz-size:"96"`
}

// New in Gloas (EIP-7732)
type PayloadAttestationData struct {
	BeaconBlockRoot   [32]byte `json:"beacon_block_root" ssz-size:"32"`
	Slot              uint64   `json:"slot"`
	PayloadPresent    bool     `json:"payload_present"`
	BlobDataAvailable bool     `json:"blob_data_available"`
}

// New in Gloas (EIP-7732)
// AggregationBits is a Bitvector[PTC_SIZE] where PTC_SIZE = 512 (64 bytes)
type PayloadAttestation struct {
	AggregationBits []byte                  `json:"aggregation_bits" ssz-size:"64"`
	Data            *PayloadAttestationData `json:"data"`
	Signature       [96]byte                `json:"signature" ssz-size:"96"`
}

// Modified in Gloas (EIP-8282): added BuilderDeposits and BuilderExits
type ExecutionRequests struct {
	Deposits        []*DepositRequest        `json:"deposits" ssz-max:"8192"`
	Withdrawals     []*WithdrawalRequest     `json:"withdrawals" ssz-max:"16"`
	Consolidations  []*ConsolidationRequest  `json:"consolidations" ssz-max:"2"`
	BuilderDeposits []*BuilderDepositRequest `json:"builder_deposits" ssz-max:"64"`
	BuilderExits    []*BuilderExitRequest    `json:"builder_exits" ssz-max:"16"`
}

type DepositRequest struct {
	Pubkey                []byte `json:"pubkey" ssz-size:"48"`
	WithdrawalCredentials []byte `json:"withdrawal_credentials" ssz-size:"32"`
	Amount                uint64 `json:"amount"`
	Signature             []byte `json:"signature" ssz-size:"96"`
	Index                 uint64 `json:"index"`
}

type WithdrawalRequest struct {
	SourceAddress   []byte `json:"source_address" ssz-size:"20"`
	ValidatorPubkey []byte `json:"validator_pubkey" ssz-size:"48"`
	Amount          uint64 `json:"amount"`
}

type ConsolidationRequest struct {
	SourceAddress []byte `json:"source_address" ssz-size:"20"`
	SourcePubkey  []byte `json:"source_pubkey" ssz-size:"48"`
	TargetPubkey  []byte `json:"target_pubkey" ssz-size:"48"`
}

// New in Gloas (EIP-8282)
type BuilderDepositRequest struct {
	Pubkey                []byte `json:"pubkey" ssz-size:"48"`
	WithdrawalCredentials []byte `json:"withdrawal_credentials" ssz-size:"32"`
	Amount                uint64 `json:"amount"`
	Signature             []byte `json:"signature" ssz-size:"96"`
}

// New in Gloas (EIP-8282)
type BuilderExitRequest struct {
	SourceAddress []byte `json:"source_address" ssz-size:"20"`
	Pubkey        []byte `json:"pubkey" ssz-size:"48"`
}

type AttesterSlashing struct {
	Attestation1 *IndexedAttestation `json:"attestation_1"`
	Attestation2 *IndexedAttestation `json:"attestation_2"`
}

type IndexedAttestation struct {
	AttestingIndices []uint64                 `json:"attesting_indices" ssz-max:"131072"`
	Data             *generic.AttestationData `json:"data"`
	Signature        []byte                   `json:"signature" ssz-size:"96"`
}

// In Gloas the execution payload is never part of the block itself.
func (b *SignedBeaconBlock) HasExecutionPayload() bool {
	return false
}

func (b *SignedBeaconBlock) Withdrawals() []*generic.Withdrawal {
	return nil
}
