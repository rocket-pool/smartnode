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
type Eth2Config struct {
    GenesisForkVersion []byte
    DomainDeposit []byte
    DomainVoluntaryExit []byte
    GenesisEpoch uint64
    GenesisTime uint64
    SecondsPerEpoch uint64
}
type BeaconHead struct {
    Slot uint64
    FinalizedSlot uint64
    JustifiedSlot uint64
    PreviousJustifiedSlot uint64
    Epoch uint64
    FinalizedEpoch uint64
    JustifiedEpoch uint64
    PreviousJustifiedEpoch uint64
}
type ValidatorStatus struct {
    Pubkey types.ValidatorPubkey
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


// Beacon client interface
type Client interface {
    GetEth2Config() (Eth2Config, error)
    GetBeaconHead() (BeaconHead, error)
    GetValidatorStatus(pubkey types.ValidatorPubkey, opts *ValidatorStatusOptions) (ValidatorStatus, error)
    Close()
}

