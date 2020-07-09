package minipool


// Minipool statuses
type MinipoolStatus int
const (
    Initialized MinipoolStatus = iota
    Prelaunch
    Staking
    Exited
    Withdrawable
    Dissolved
)
func (s MinipoolStatus) String() string {
    return []string{"Initialized", "Prelaunch", "Staking", "Exited", "Withdrawable", "Dissolved"}[s]
}


// Minipool deposit types
type MinipoolDeposit int
const (
    None MinipoolDeposit = iota
    Full
    Half
    Empty
)
func (d MinipoolDeposit) String() string {
    return []string{"None", "Full", "Half", "Empty"}[d]
}

