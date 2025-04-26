package eth2

import (
	"fmt"
	"strings"
)

const SlotsPerHistoricalRoot uint64 = 8192

// Deposit data (with no signature field)
type DepositDataNoSignature struct {
	PublicKey             []byte `json:"pubkey" ssz-size:"48"`
	WithdrawalCredentials []byte `json:"withdrawal_credentials" ssz-size:"32"`
	Amount                uint64 `json:"amount"`
}

// Deposit data (including signature)
type DepositData struct {
	PublicKey             []byte `json:"pubkey" ssz-size:"48"`
	WithdrawalCredentials []byte `json:"withdrawal_credentials" ssz-size:"32"`
	Amount                uint64 `json:"amount"`
	Signature             []byte `json:"signature" ssz-size:"96"`
}

// BLS signing root with domain
type SigningRoot struct {
	ObjectRoot []byte `json:"object_root" ssz-size:"32"`
	Domain     []byte `json:"domain" ssz-size:"32"`
}

// Voluntary exit transaction
type VoluntaryExit struct {
	Epoch          uint64 `json:"epoch"`
	ValidatorIndex uint64 `json:"validator_index"`
}

// Withdrawal creds change message
type WithdrawalCredentialsChange struct {
	ValidatorIndex     uint64   `json:"validator_index"`
	FromBLSPubkey      [48]byte `json:"from_bls_pubkey" ssz-size:"48"`
	ToExecutionAddress [20]byte `json:"to_execution_address" ssz-size:"20"`
}

type BeaconState interface {
	GetSlot() uint64
	ValidatorWithdrawableEpochProof(index uint64) ([][]byte, error)
	ValidatorCredentialsProof(index uint64) ([][]byte, error)
	HistoricalSummaryProof(slot uint64) ([][]byte, error)
	HistoricalSummaryBlockRootProof(slot int) ([][]byte, error)
	BlockRootProof(slot uint64) ([][]byte, error)
	GetValidators() []*Validator
}

type BeaconBlock interface {
	ProveWithdrawal(indexInWithdrawalsArray uint64) ([][]byte, error)
	HasExecutionPayload() bool
	Withdrawals() []*Withdrawal
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

type HistoricalSummaryLists struct {
	BlockRoots [8192][32]byte `json:"block_roots" ssz-size:"8192,32"`
	StateRoots [8192][32]byte `json:"state_roots" ssz-size:"8192,32"`
}

func NewBeaconState(data []byte, fork string) (BeaconState, error) {
	fork = strings.ToLower(fork)

	switch fork {
	case "deneb":
		out := &BeaconStateDeneb{}
		err := out.UnmarshalSSZ(data)
		if err != nil {
			return nil, err
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported fork: %s", fork)
	}
}

func NewBeaconBlock(data []byte, fork string) (BeaconBlock, error) {
	fork = strings.ToLower(fork)

	switch fork {
	case "deneb":
		out := &BeaconBlockDeneb{}
		err := out.UnmarshalSSZ(data)
		if err != nil {
			return nil, err
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported fork: %s", fork)
	}
}
