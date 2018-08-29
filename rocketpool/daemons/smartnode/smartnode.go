package smartnode

import ()


// Config
const checkRPIPVoteInterval string = "10s"


// Run daemon
func Run() {

    // Check for node exit on minipool removed
    go startCheckNodeExit()
    go checkNodeExit()

    // Check for RPIP alerts on new proposal
    go startCheckRPIPAlerts()
    go checkRPIPAlerts()

    // Check RPIP votes periodically
    go startCheckRPIPVotes(checkRPIPVoteInterval)
    go checkRPIPVotes()

    // Block thread
    select {}

}

