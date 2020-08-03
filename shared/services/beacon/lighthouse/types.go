package lighthouse

import (
    "encoding/hex"
    "encoding/json"

    hexutil "github.com/rocket-pool/smartnode/shared/utils/hex"
)


// Request types
type ValidatorsRequest struct {
    StateRoot string                `json:"state_root,omitempty"`
    Pubkeys []string                `json:"pubkeys"`
}

// Response types
type Eth2ConfigResponse struct {
    GenesisForkVersion byteArray    `json:"genesis_fork_version"`
    DomainDeposit uint64            `json:"domain_deposit"`
    DomainVoluntaryExit uint64      `json:"domain_voluntary_exit"`
    GenesisSlot uint64              `json:"genesis_slot"`
    MillisecondsPerSlot uint64      `json:"milliseconds_per_slot"`
}
type BeaconHeadResponse struct {
    Slot uint64                     `json:"slot"`
    FinalizedSlot uint64            `json:"finalized_slot"`
    JustifiedSlot uint64            `json:"justified_slot"`
    PreviousJustifiedSlot uint64    `json:"previous_justified_slot"`
}
type ValidatorResponse struct {
    Balance uint64                  `json:"balance"`
    Validator struct {
        Pubkey byteArray                    `json:"pubkey"`
        WithdrawalCredentials byteArray     `json:"withdrawal_credentials"`
        EffectiveBalance uint64             `json:"effective_balance"`
        Slashed bool                        `json:"slashed"`
        ActivationEligibilityEpoch uint64   `json:"activation_eligibility_epoch"`
        ActivationEpoch uint64              `json:"activation_epoch"`
        ExitEpoch uint64                    `json:"exit_epoch"`
        WithdrawableEpoch uint64            `json:"withdrawable_epoch"`
    }                               `json:"validator"`
}


// Byte array
type byteArray []byte


// JSON encoding
func (b *byteArray) UnmarshalJSON(data []byte) error {

    // Unmarshal string
    var dataStr string
    if err := json.Unmarshal(data, &dataStr); err != nil {
        return err
    }

    // Decode hex
    value, err := hex.DecodeString(hexutil.RemovePrefix(dataStr))
    if err != nil {
        return err
    }

    // Set value and return
    *b = value
    return nil

}

