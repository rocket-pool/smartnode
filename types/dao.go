package types

import (
    "encoding/json"
    "fmt"
)


// DAO proposal states
type ProposalState uint8
const (
    Pending ProposalState = iota
    Active
    Cancelled
    Defeated
    Succeeded
    Expired
    Executed
)
var ProposalStates = []string{"Pending", "Active", "Cancelled", "Defeated", "Succeeded", "Expired", "Executed"}


// String conversion
func (s ProposalState) String() string {
    if int(s) >= len(ProposalStates) { return "" }
    return ProposalStates[s]
}
func StringToProposalState(value string) (ProposalState, error) {
    for state, str := range ProposalStates {
        if value == str { return ProposalState(state), nil }
    }
    return 0, fmt.Errorf("Invalid proposal state '%s'", value)
}


// JSON encoding
func (s ProposalState) MarshalJSON() ([]byte, error) {
    str := s.String()
    if str == "" {
        return []byte{}, fmt.Errorf("Invalid proposal state '%d'", s)
    }
    return json.Marshal(str)
}
func (s *ProposalState) UnmarshalJSON(data []byte) error {
    var dataStr string
    if err := json.Unmarshal(data, &dataStr); err != nil { return err }
    state, err := StringToProposalState(dataStr)
    if err == nil { *s = state }
    return err
}

