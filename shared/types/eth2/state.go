package eth2

import (
	"bytes"
	"errors"
	"fmt"
	"math/bits"
	"reflect"
	"sync/atomic"
)

// Important indices for proof generation:
// BeaconStateDenebValidatorsIndex is the field offset of the Validators field in the BeaconStateDeneb struct
const beaconStateDenebValidatorsIndex uint64 = 11

// If this ever isn't a power of two, we need to round up to the next power of two
const beaconStateValidatorsMaxLength uint64 = 1 << 40
const beaconBlockHeaderStateRootGeneralizedIndex uint64 = 11                     // Container with 5 fields, so gid 8 is the first field. We want the 4th field, so gid 8 + 3 = 11
const beaconStateValidatorWithdrawalCredentialsPubkeyGeneralizedIndex uint64 = 4 // Container with 8 fields, so gid 8 is the first field. We want the parent of 1st field, so gid 8 / 2 = 4
const beaconStateValidatorWithdrawableEpochGeneralizedIndex uint64 = 14          // Container with 8 fields, so gid 8 is the first field. We want the 8th field, so gid 8 + 7 = 15
// See https://github.com/ethereum/consensus-specs/blob/dev/ssz/merkle-proofs.md for general index calculation and helpers

func getPowerOfTwoCeil(x uint64) uint64 {
	// Base case
	if x <= 1 {
		return 1
	}

	// Check if already a power of two
	if x&(x-1) == 0 {
		return x
	}

	// Find the most significant bit
	msb := bits.Len64(x) - 1
	return 1 << (msb + 1)
}

var beaconStateChunkSize atomic.Uint64

func getDenebStateChunkSize() uint64 {
	// Use a static value to avoid multiple reflection calls
	storedChunkSize := beaconStateChunkSize.Load()
	if storedChunkSize == 0 {
		s := reflect.TypeOf(BeaconStateDeneb{}).NumField()
		beaconStateChunkSize.Store(uint64(s))
		storedChunkSize = uint64(s)
	}
	return storedChunkSize
}

func getDenebGeneralizedIndexForValidators() uint64 {
	// There's 28 fields, so rounding up to the next power of two is 32, a left-aligned node
	// BeaconStateDenebValidatorsIndex is the 11th field, so its generalized index is 32 + 11 = 43
	return getPowerOfTwoCeil(getDenebStateChunkSize()) + beaconStateDenebValidatorsIndex
}

func (state *BeaconStateDeneb) getGeneralizedIndexForValidator(index uint64) uint64 {
	root := getDenebGeneralizedIndexForValidators()

	// Now, grab the validator index within the list
	// `start` is `index * 32` and `pos` is `start / 32` so pos is just `index`
	pos := index
	baseIndex := uint64(2) // Lists have a base index of 2
	root = root*baseIndex*getPowerOfTwoCeil(beaconStateValidatorsMaxLength) + pos

	// root is now the generalized index for the validator
	return root
}

func (state *BeaconStateDeneb) validatorStateProof(index uint64) ([][]byte, error) {

	// Convert the state to a proof tree
	root, err := state.GetTree()
	if err != nil {
		return nil, fmt.Errorf("could not get state tree: %w", err)
	}

	// Find the validator's generalized index
	generalizedIndex := state.getGeneralizedIndexForValidator(index)

	// Grab the proof for that index
	proof, err := root.Prove(int(generalizedIndex))
	if err != nil {
		return nil, fmt.Errorf("could not get proof for validator: %w", err)
	}

	// Sanity check that the proof leaf matches the expected validator
	validatorHashTreeRoot, err := state.Validators[index].HashTreeRoot()
	if err != nil {
		return nil, fmt.Errorf("could not get hash tree root for validator: %w", err)
	}
	if !bytes.Equal(proof.Leaf, validatorHashTreeRoot[:]) {
		return nil, fmt.Errorf("proof leaf does not match expected validator")
	}

	return proof.Hashes, nil

}

func (validator *Validator) validatorWithdrawableEpochProof() ([][]byte, error) {
	// Just get the portion of the proof for the validator ExitEpoch.
	generalizedIndex := beaconStateValidatorWithdrawableEpochGeneralizedIndex
	root, err := validator.GetTree()
	if err != nil {
		return nil, fmt.Errorf("could not get validator tree: %w", err)
	}
	proof, err := root.Prove(int(generalizedIndex))
	if err != nil {
		return nil, fmt.Errorf("could not get proof for validator withdrawable epoch: %w", err)
	}
	return proof.Hashes, nil
}

// ValidatorWithdrawableEpochProof computes the merkle proof for a validator's withdrawable epoch
// at a specific index in the validator registry.
func (state *BeaconStateDeneb) ValidatorWithdrawableEpochProof(index uint64) ([][]byte, error) {

	if index >= uint64(len(state.Validators)) {
		return nil, errors.New("validator index out of bounds")
	}

	// Get the validator's withdrawable epoch proof
	withdrawableEpochProof, err := state.Validators[index].validatorWithdrawableEpochProof()
	if err != nil {
		return nil, fmt.Errorf("could not get validator withdrawable epoch proof: %w", err)
	}

	stateProof, err := state.validatorStateProof(index)
	if err != nil {
		return nil, fmt.Errorf("could not get validator state proof: %w", err)
	}

	// The EL proves against BeaconBlockHeader root, so we need to merge the state proof with that.
	generalizedIndex := beaconBlockHeaderStateRootGeneralizedIndex
	root, err := state.LatestBlockHeader.GetTree()
	if err != nil {
		return nil, fmt.Errorf("could not get block header tree: %w", err)
	}
	blockHeaderProof, err := root.Prove(int(generalizedIndex))
	if err != nil {
		return nil, fmt.Errorf("could not get proof for block header: %w", err)
	}

	out := append(withdrawableEpochProof, stateProof...)
	out = append(out, blockHeaderProof.Hashes...)

	return out, nil
}

func (validator *Validator) validatorCredentialsPubkeyProof() ([][]byte, error) {
	// Just get the portion of the proof for the validator's credentials.
	generalizedIndex := beaconStateValidatorWithdrawalCredentialsPubkeyGeneralizedIndex
	root, err := validator.GetTree()
	if err != nil {
		return nil, fmt.Errorf("could not get validator tree: %w", err)
	}
	proof, err := root.Prove(int(generalizedIndex))
	if err != nil {
		return nil, fmt.Errorf("could not get proof for validator credentials: %w", err)
	}
	return proof.Hashes, nil
}

// ValidatorCredentialsProof computes the merkle proof for a validator's credentials
// at a specific index in the validator registry.
func (state *BeaconStateDeneb) ValidatorCredentialsProof(index uint64) ([][]byte, error) {

	if index >= uint64(len(state.Validators)) {
		return nil, errors.New("validator index out of bounds")
	}

	// Get the validator's credentials proof
	credentialsProof, err := state.Validators[index].validatorCredentialsPubkeyProof()
	if err != nil {
		return nil, fmt.Errorf("could not get validator credentials proof: %w", err)
	}

	stateProof, err := state.validatorStateProof(index)
	if err != nil {
		return nil, fmt.Errorf("could not get validator state proof: %w", err)
	}

	// The EL proves against BeaconBlockHeader root, so we need to merge the state proof with that.
	generalizedIndex := beaconBlockHeaderStateRootGeneralizedIndex
	root, err := state.LatestBlockHeader.GetTree()
	if err != nil {
		return nil, fmt.Errorf("could not get block header tree: %w", err)
	}
	blockHeaderProof, err := root.Prove(int(generalizedIndex))
	if err != nil {
		return nil, fmt.Errorf("could not get proof for block header: %w", err)
	}

	out := append(credentialsProof, stateProof...)
	out = append(out, blockHeaderProof.Hashes...)

	return out, nil
}

// Taken from https://github.com/prysmaticlabs/prysm/blob/ac1717f1e44bd218b0bd3af0c4dec951c075f462/proto/prysm/v1alpha1/beacon_state.pb.go#L1574
// Unexported fields stripped, as well as proto-related field tags. JSON and ssz-size tags are preserved, and nested types are replaced with local copies as well.
type BeaconStateDeneb struct {
	GenesisTime                  uint64                       `json:"genesis_time"`
	GenesisValidatorsRoot        []byte                       `json:"genesis_validators_root" ssz-size:"32"`
	Slot                         uint64                       `json:"slot"`
	Fork                         *Fork                        `json:"fork"`
	LatestBlockHeader            *BeaconBlockHeader           `json:"latest_block_header"`
	BlockRoots                   [8192][32]byte               `json:"block_roots" ssz-size:"8192,32"`
	StateRoots                   [8192][32]byte               `json:"state_roots" ssz-size:"8192,32"`
	HistoricalRoots              [][]byte                     `json:"historical_roots" ssz-max:"16777216" ssz-size:"?,32"`
	Eth1Data                     *Eth1Data                    `json:"eth1_data"`
	Eth1DataVotes                []*Eth1Data                  `json:"eth1_data_votes" ssz-max:"2048"`
	Eth1DepositIndex             uint64                       `json:"eth1_deposit_index"`
	Validators                   []*Validator                 `json:"validators" ssz-max:"1099511627776"`
	Balances                     []uint64                     `json:"balances" ssz-max:"1099511627776"`
	RandaoMixes                  [][]byte                     `json:"randao_mixes" ssz-size:"65536,32"`
	Slashings                    []uint64                     `json:"slashings" ssz-size:"8192"`
	PreviousEpochParticipation   []byte                       `json:"previous_epoch_participation" ssz-max:"1099511627776"`
	CurrentEpochParticipation    []byte                       `json:"current_epoch_participation" ssz-max:"1099511627776"`
	JustificationBits            [1]byte                      `json:"justification_bits" ssz-size:"1"`
	PreviousJustifiedCheckpoint  *Checkpoint                  `json:"previous_justified_checkpoint"`
	CurrentJustifiedCheckpoint   *Checkpoint                  `json:"current_justified_checkpoint"`
	FinalizedCheckpoint          *Checkpoint                  `json:"finalized_checkpoint"`
	InactivityScores             []uint64                     `json:"inactivity_scores" ssz-max:"1099511627776"`
	CurrentSyncCommittee         *SyncCommittee               `json:"current_sync_committee"`
	NextSyncCommittee            *SyncCommittee               `json:"next_sync_committee"`
	LatestExecutionPayloadHeader *ExecutionPayloadHeaderDeneb `json:"latest_execution_payload_header"`
	NextWithdrawalIndex          uint64                       `json:"next_withdrawal_index"`
	NextWithdrawalValidatorIndex uint64                       `json:"next_withdrawal_validator_index"`
	HistoricalSummaries          []*HistoricalSummary         `json:"historical_summaries" ssz-max:"16777216"`
}

// Remaining types taken from https://github.com/ferranbt/fastssz/blob/03cd29050aa2555fd4abc29ace7c1fac8b8fb25e/spectests/structs.go

// Per-Fork ExecutionPayloadHeaders

type ExecutionPayloadHeaderDeneb struct {
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

// Generic types

type Uint256 [32]byte

type Fork struct {
	PreviousVersion []byte `json:"previous_version" ssz-size:"4"`
	CurrentVersion  []byte `json:"current_version" ssz-size:"4"`
	Epoch           uint64 `json:"epoch"`
}

type BeaconBlockHeader struct {
	Slot          uint64 `json:"slot"`
	ProposerIndex uint64 `json:"proposer_index"`
	ParentRoot    []byte `json:"parent_root" ssz-size:"32"`
	StateRoot     []byte `json:"state_root" ssz-size:"32"`
	BodyRoot      []byte `json:"body_root" ssz-size:"32"`
}

type Eth1Data struct {
	DepositRoot  []byte `json:"deposit_root" ssz-size:"32"`
	DepositCount uint64 `json:"deposit_count"`
	BlockHash    []byte `json:"block_hash" ssz-size:"32"`
}

type Validator struct {
	Pubkey                     []byte `json:"pubkey" ssz-size:"48"`
	WithdrawalCredentials      []byte `json:"withdrawal_credentials" ssz-size:"32"`
	EffectiveBalance           uint64 `json:"effective_balance"`
	Slashed                    bool   `json:"slashed"`
	ActivationEligibilityEpoch uint64 `json:"activation_eligibility_epoch"`
	ActivationEpoch            uint64 `json:"activation_epoch"`
	ExitEpoch                  uint64 `json:"exit_epoch"`
	WithdrawableEpoch          uint64 `json:"withdrawable_epoch"`
}

type Checkpoint struct {
	Epoch uint64 `json:"epoch"`
	Root  []byte `json:"root" ssz-size:"32"`
}

type SyncCommittee struct {
	PubKeys         [][]byte `json:"pubkeys" ssz-size:"512,48"`
	AggregatePubKey [48]byte `json:"aggregate_pubkey" ssz-size:"48"`
}

type HistoricalSummary struct {
	BlockSummaryRoot [32]byte `json:"block_summary_root" ssz-size:"32"`
	StateSummaryRoot [32]byte `json:"state_summary_root" ssz-size:"32"`
}
