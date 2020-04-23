package beacon


// API response types
type Eth2Config struct {
    GenesisForkVersion []byte
    BLSWithdrawalPrefixByte byte
    DomainBeaconProposer []byte
    DomainBeaconAttester []byte
    DomainRandao []byte
    DomainDeposit []byte
    DomainVoluntaryExit []byte
    SlotsPerEpoch uint64
}
type BeaconHead struct {
    Epoch uint64
    FinalizedEpoch uint64
    JustifiedEpoch uint64
}
type ValidatorStatus struct {
    Pubkey []byte
    WithdrawalCredentials []byte
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
    GetEth2Config() (*Eth2Config, error)
    GetBeaconHead() (*BeaconHead, error)
    GetValidatorStatus(pubkey []byte) (*ValidatorStatus, error)
}

