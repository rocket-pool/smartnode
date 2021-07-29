package nimbus

import (
	"encoding/hex"
	"encoding/json"
	"strconv"

	hexutil "github.com/rocket-pool/smartnode/shared/utils/hex"
)

// Request types
type VoluntaryExitRequest struct {
    Message   VoluntaryExitMessage `json:"message"`
    Signature byteArray            `json:"signature"`
}
type VoluntaryExitMessage struct {
    Epoch          uint64 `json:"epoch"`
    ValidatorIndex uint64 `json:"validator_index"`
}

// Response types
type SyncStatusResponse struct {
    IsSyncing bool                      `json:"is_syncing"`
    HeadSlot uint64                     `json:"head_slot"`
    SyncDistance uint64                 `json:"sync_distance"`
}
type Eth2ConfigResponse struct {
    SecondsPerSlot uinteger `json:"SECONDS_PER_SLOT"`
    SlotsPerEpoch  uinteger `json:"SLOTS_PER_EPOCH"`
}
type GenesisResponse struct {
    GenesisTime           uint64    `json:"genesis_time"`
    GenesisForkVersion    byteArray `json:"genesis_fork_version"`
    GenesisValidatorsRoot byteArray `json:"genesis_validators_root"`
}
type FinalityCheckpointsResponse struct {
    PreviousJustified struct {
        Epoch uint64 `json:"epoch"`
    } `json:"previous_justified"`
    CurrentJustified struct {
        Epoch uint64 `json:"epoch"`
    } `json:"current_justified"`
    Finalized struct {
        Epoch uint64 `json:"epoch"`
    } `json:"finalized"`
}
type ForkResponse struct {
    PreviousVersion byteArray `json:"previous_version"`
    CurrentVersion  byteArray `json:"current_version"`
    Epoch           uint64    `json:"epoch"`
}

type Validator struct {
    Index     uint64 `json:"index"`
    Balance   uint64 `json:"balance"`
    Status    string `json:"status"`
    Validator struct {
        Pubkey                     byteArray `json:"pubkey"`
        WithdrawalCredentials      byteArray `json:"withdrawal_credentials"`
        EffectiveBalance           uint64    `json:"effective_balance"`
        Slashed                    bool      `json:"slashed"`
        ActivationEligibilityEpoch int64     `json:"activation_eligibility_epoch"`  // Nimbus uses -1 for FAR_FUTURE_EPOCH so this has to be a signed int
        ActivationEpoch            int64     `json:"activation_epoch"`              // Same here
        ExitEpoch                  int64     `json:"exit_epoch"`                    // Same here
        WithdrawableEpoch          int64     `json:"withdrawable_epoch"`            // Same here
    } `json:"validator"`
}

// Unsigned integer type
type uinteger uint64

func (i uinteger) MarshalJSON() ([]byte, error) {
    return json.Marshal(strconv.Itoa(int(i)))
}
func (i *uinteger) UnmarshalJSON(data []byte) error {

    // Unmarshal string
    var dataStr string
    if err := json.Unmarshal(data, &dataStr); err != nil {
        return err
    }

    // Parse integer value
    value, err := strconv.ParseUint(dataStr, 10, 64)
    if err != nil {
        return err
    }

    // Set value and return
    *i = uinteger(value)
    return nil

}

// Byte array type
type byteArray []byte

func (b byteArray) MarshalJSON() ([]byte, error) {
    return json.Marshal(hexutil.AddPrefix(hex.EncodeToString(b)))
}
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
