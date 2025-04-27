package electra

import "github.com/rocket-pool/smartnode/shared/types/eth2/generic"

// Important indices for proof generation:
const BeaconBlockElectraBodyChunksCeil uint64 = 16

func (b *BeaconBlockElectra) ProveWithdrawal(indexInWithdrawalsArray uint64) ([][]byte, error) {
	tree, err := b.GetTree()
	if err != nil {
		return nil, err
	}

	gid := uint64(1)
	// Navigate to the body
	gid = gid*generic.BeaconBlockChunksCeil + generic.BeaconBlockBodyIndex
	// Then to the ExecutionPayload
	gid = gid*BeaconBlockElectraBodyChunksCeil + generic.BeaconBlockBodyExecutionPayloadIndex
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
type BeaconBlockElectra struct {
	Slot          uint64                  `json:"slot"`
	ProposerIndex uint64                  `json:"proposer_index"`
	ParentRoot    [32]byte                `json:"parent_root" ssz-size:"32"`
	StateRoot     [32]byte                `json:"state_root" ssz-size:"32"`
	Body          *BeaconBlockBodyElectra `json:"body"`
}

type SignedBeaconBlockElectra struct {
	Block     *BeaconBlockElectra `json:"message"`
	Signature []byte              `json:"signature" ssz-size:"96"`
}

type BeaconBlockBodyElectra struct {
	RandaoReveal          []byte                                `json:"randao_reveal" ssz-size:"96"`
	Eth1Data              *generic.Eth1Data                     `json:"eth1_data"`
	Graffiti              [32]byte                              `json:"graffiti" ssz-size:"32"`
	ProposerSlashings     []*generic.ProposerSlashing           `json:"proposer_slashings" ssz-max:"16"`
	AttesterSlashings     []*AttesterSlashingElectra            `json:"attester_slashings" ssz-max:"1"`
	Attestations          []*generic.Attestation                `json:"attestations" ssz-max:"8"`
	Deposits              []*generic.Deposit                    `json:"deposits" ssz-max:"16"`
	VoluntaryExits        []*generic.SignedVoluntaryExit        `json:"voluntary_exits" ssz-max:"16"`
	SyncAggregate         *generic.SyncAggregate                `json:"sync_aggregate"`
	ExecutionPayload      *generic.ExecutionPayload             `json:"execution_payload"`
	BlsToExecutionChanges []*generic.SignedBLSToExecutionChange `json:"bls_to_execution_changes" ssz-max:"16"`
	BlobKzgCommitments    [][48]byte                            `json:"blob_kzg_commitments" ssz-max:"4096,48" ssz-size:"?,48"`
	ExecutionRequests     *ExecutionRequests                    `json:"execution_requests"`
}

type ExecutionRequests struct {
	Deposits       []*DepositRequest       `json:"deposits" ssz-max:"8192"`
	Withdrawals    []*WithdrawalRequest    `json:"withdrawals" ssz-max:"16"`
	Consolidations []*ConsolidationRequest `json:"consolidations" ssz-max:"2"`
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

type AttesterSlashingElectra struct {
	Attestation1 *generic.IndexedAttestation `json:"attestation_1"`
	Attestation2 *generic.IndexedAttestation `json:"attestation_2"`
}

func (b *BeaconBlockElectra) HasExecutionPayload() bool {
	return b.Body.ExecutionPayload != nil
}

func (b *BeaconBlockElectra) Withdrawals() []*generic.Withdrawal {
	return b.Body.ExecutionPayload.Withdrawals
}
