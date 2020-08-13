package types

import (
    "encoding/json"
    "fmt"
)


// Minipool statuses
type MinipoolStatus uint8
const (
    Initialized MinipoolStatus = iota
    Prelaunch
    Staking
    Withdrawable
    Dissolved
)
var MinipoolStatuses = []string{"Initialized", "Prelaunch", "Staking", "Withdrawable", "Dissolved"}


// String conversion
func (s MinipoolStatus) String() string {
    if int(s) >= len(MinipoolStatuses) { return "" }
    return MinipoolStatuses[s]
}
func StringToMinipoolStatus(value string) (MinipoolStatus, error) {
    for status, str := range MinipoolStatuses {
        if value == str { return MinipoolStatus(status), nil }
    }
    return 0, fmt.Errorf("Invalid minipool status '%s'", value)
}


// JSON encoding
func (s MinipoolStatus) MarshalJSON() ([]byte, error) {
    str := s.String()
    if str == "" {
        return []byte{}, fmt.Errorf("Invalid minipool status '%d'", s)
    }
    return json.Marshal(str)
}
func (s *MinipoolStatus) UnmarshalJSON(data []byte) error {
    var dataStr string
    if err := json.Unmarshal(data, &dataStr); err != nil { return err }
    status, err := StringToMinipoolStatus(dataStr)
    if err == nil { *s = status }
    return err
}


// Minipool deposit types
type MinipoolDeposit uint8
const (
    None MinipoolDeposit = iota
    Full
    Half
    Empty
)
var MinipoolDepositTypes = []string{"None", "Full", "Half", "Empty"}


// String conversion
func (d MinipoolDeposit) String() string {
    if int(d) >= len(MinipoolDepositTypes) { return "" }
    return MinipoolDepositTypes[d]
}
func StringToMinipoolDeposit(value string) (MinipoolDeposit, error) {
    for depositType, str := range MinipoolDepositTypes {
        if value == str { return MinipoolDeposit(depositType), nil }
    }
    return 0, fmt.Errorf("Invalid minipool deposit type '%s'", value)
}


// JSON encoding
func (d MinipoolDeposit) MarshalJSON() ([]byte, error) {
    str := d.String()
    if str == "" {
        return []byte{}, fmt.Errorf("Invalid minipool deposit type '%d'", d)
    }
    return json.Marshal(str)
}
func (d *MinipoolDeposit) UnmarshalJSON(data []byte) error {
    var dataStr string
    if err := json.Unmarshal(data, &dataStr); err != nil { return err }
    depositType, err := StringToMinipoolDeposit(dataStr)
    if err == nil { *d = depositType }
    return err
}

