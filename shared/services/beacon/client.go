package beacon

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/prysmaticlabs/go-bitfield"
	"github.com/rocket-pool/rocketpool-go/types"
)

// API request options
type ValidatorStatusOptions struct {
	Epoch *uint64
	Slot  *uint64
}

// API response types
type SyncStatus struct {
	Syncing  bool
	Progress float64
}
type Eth2DepositContract struct {
	ChainID uint64
	Address common.Address
}
type BeaconHead struct {
	Epoch                  uint64
	FinalizedEpoch         uint64
	JustifiedEpoch         uint64
	PreviousJustifiedEpoch uint64
}
type ValidatorStatus struct {
	Pubkey                     types.ValidatorPubkey `json:"pubkey"`
	Index                      string                `json:"index"`
	WithdrawalCredentials      common.Hash           `json:"withdrawal_credentials"`
	Balance                    uint64                `json:"balance"`
	Status                     ValidatorState        `json:"status"`
	EffectiveBalance           uint64                `json:"effective_balance"`
	Slashed                    bool                  `json:"slashed"`
	ActivationEligibilityEpoch uint64                `json:"activation_eligibility_epoch"`
	ActivationEpoch            uint64                `json:"activation_epoch"`
	ExitEpoch                  uint64                `json:"exit_epoch"`
	WithdrawableEpoch          uint64                `json:"withdrawable_epoch"`
	Exists                     bool                  `json:"exists"`
}
type Eth1Data struct {
	DepositRoot  common.Hash
	DepositCount uint64
	BlockHash    common.Hash
}
type BeaconBlock struct {
	Slot                 uint64
	ProposerIndex        string
	HasExecutionPayload  bool
	Attestations         []AttestationInfo
	FeeRecipient         common.Address
	ExecutionBlockNumber uint64
}
type BeaconBlockHeader struct {
	Slot          uint64
	ProposerIndex string
}

// Committees is an interface as an optimization- since committees responses
// are quite large, there's a decent cpu/memory improvement to removing the
// translation to an intermediate storage class.
//
// Instead, the interface provides the access pattern that smartnode (or more
// specifically, tree-gen) wants, and the underlying format is just the format
// of the Beacon Node response.
type Committees interface {
	// Index returns the index of the committee at the provided offset
	Index(int) uint64
	// Slot returns the slot of the committee at the provided offset
	Slot(int) uint64
	// Validators returns the list of validators of the committee at
	// the provided offset
	Validators(int) []string
	// Count returns the number of committees in the response
	Count() int
	// Release returns the reused validators slice buffer to the pool for
	// further reuse, and must be called when the user is done with this
	// committees instance
	Release()
}

type AttestationInfo struct {
	AggregationBits bitfield.Bitlist
	SlotIndex       uint64
	CommitteeIndex  uint64
}

// Beacon client type
type BeaconClientType int

const (
	// This client is a traditional "split process" design, where the beacon
	// client and validator process are separate and run in different
	// containers
	SplitProcess BeaconClientType = iota

	// This client is a "single process" where the beacon client and
	// validator run in the same process (or run as separate processes
	// within the same docker container)
	SingleProcess

	// Unknown / missing client type
	Unknown
)

type ValidatorState string

const (
	ValidatorState_PendingInitialized ValidatorState = "pending_initialized"
	ValidatorState_PendingQueued      ValidatorState = "pending_queued"
	ValidatorState_ActiveOngoing      ValidatorState = "active_ongoing"
	ValidatorState_ActiveExiting      ValidatorState = "active_exiting"
	ValidatorState_ActiveSlashed      ValidatorState = "active_slashed"
	ValidatorState_ExitedUnslashed    ValidatorState = "exited_unslashed"
	ValidatorState_ExitedSlashed      ValidatorState = "exited_slashed"
	ValidatorState_WithdrawalPossible ValidatorState = "withdrawal_possible"
	ValidatorState_WithdrawalDone     ValidatorState = "withdrawal_done"
)

// Beacon client interface
type Client interface {
	GetClientType() (BeaconClientType, error)
	GetSyncStatus() (SyncStatus, error)
	GetEth2Config() (Eth2Config, error)
	GetEth2DepositContract() (Eth2DepositContract, error)
	GetAttestations(blockId string) ([]AttestationInfo, bool, error)
	GetBeaconBlock(blockId string) (BeaconBlock, bool, error)
	GetBeaconBlockHeader(blockId string) (BeaconBlockHeader, bool, error)
	GetBeaconHead() (BeaconHead, error)
	GetValidatorStatusByIndex(index string, opts *ValidatorStatusOptions) (ValidatorStatus, error)
	GetValidatorStatus(pubkey types.ValidatorPubkey, opts *ValidatorStatusOptions) (ValidatorStatus, error)
	GetValidatorStatuses(pubkeys []types.ValidatorPubkey, opts *ValidatorStatusOptions) (map[types.ValidatorPubkey]ValidatorStatus, error)
	GetValidatorIndex(pubkey types.ValidatorPubkey) (string, error)
	GetValidatorSyncDuties(indices []string, epoch uint64) (map[string]bool, error)
	GetValidatorProposerDuties(indices []string, epoch uint64) (map[string]uint64, error)
	GetDomainData(domainType []byte, epoch uint64, useGenesisFork bool) ([]byte, error)
	ExitValidator(validatorIndex string, epoch uint64, signature types.ValidatorSignature) error
	Close() error
	GetEth1DataForEth2Block(blockId string) (Eth1Data, bool, error)
	GetCommitteesForEpoch(epoch *uint64) (Committees, error)
	ChangeWithdrawalCredentials(validatorIndex string, fromBlsPubkey types.ValidatorPubkey, toExecutionAddress common.Address, signature types.ValidatorSignature) error
}
