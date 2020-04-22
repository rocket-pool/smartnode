package beacon


// API response types
type Eth2Config struct {
    GenesisForkVersion []byte
    BLSWithdrawalPrefixByte []byte
    DomainBeaconProposer uint64
    DomainBeaonAttester uint64
    DomainRandao uint64
    DomainDeposit uint64
    DomainVoluntaryExit uint64
    SlotsPerEpoch uint64
}
type BeaconHead struct {
    Epoch uint64
    FinalizedEpoch uint64
    JustifiedEpoch uint64
}
type ValidatorStatus struct {
    Pubkey []byte
    ValidatorIndex uint64
    Balance uint64
    EffectiveBalance uint64
    WithdrawalCredentials []byte
    Slashed bool
    ActivationEligibilityEpoch uint64
    ActivationEpoch uint64
    ExitEpoch uint64
    WithdrawableEpoch uint64
    Exists bool
}


// Beaon client interface
type Client interface {
    GetEth2Config() (*Eth2Config, error)
    GetBeaconHead() (*BeaconHead, error)
    GetValidatorStatus(pubkey string) (*ValidatorStatus, error)
}

