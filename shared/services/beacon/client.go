package beacon

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/types"
)

// API request options
type ValidatorStatusOptions struct {
    Epoch uint64
}


// API response types
type SyncStatus struct {
    Syncing bool
}
type Eth2Config struct {
    GenesisForkVersion []byte
    GenesisValidatorsRoot []byte
    GenesisEpoch uint64
    GenesisTime uint64
    SecondsPerEpoch uint64
}
type BeaconHead struct {
    Epoch uint64
    FinalizedEpoch uint64
    JustifiedEpoch uint64
    PreviousJustifiedEpoch uint64
}
type ValidatorStatus struct {
    Pubkey types.ValidatorPubkey
	ValidatorIndex uint64
    WithdrawalCredentials common.Hash
    Balance uint64
    EffectiveBalance uint64
    Slashed bool
    ActivationEligibilityEpoch uint64
    ActivationEpoch uint64
    ExitEpoch uint64
    WithdrawableEpoch uint64
    Exists bool
}


// Beacon client type
type BeaconClientType int
const(
	// This client is a traditional "split process" design, where the beacon
	// client and validator process are separate and run in different
	// containers
	SplitProcess BeaconClientType = iota

	// This client is a "single process" where the beacon client and
	// validator run in the same process (or run as separate processes
	// within the same docker container)
	SingleProcess
)


// Beacon client interface
type Client interface {
    GetClientType() (BeaconClientType)
    GetSyncStatus() (SyncStatus, error)
    GetEth2Config() (Eth2Config, error)
    GetBeaconHead() (BeaconHead, error)
    GetValidatorStatus(pubkey types.ValidatorPubkey, opts *ValidatorStatusOptions) (ValidatorStatus, error)
    GetValidatorStatuses(pubkeys []types.ValidatorPubkey, opts *ValidatorStatusOptions) (map[types.ValidatorPubkey]ValidatorStatus, error)
    GetValidatorIndex(pubkey types.ValidatorPubkey) (uint64, error)
    GetDomainData(domainType []byte, epoch uint64) ([]byte, error)
    ExitValidator(validatorIndex, epoch uint64, signature types.ValidatorSignature) error
    Close()
}