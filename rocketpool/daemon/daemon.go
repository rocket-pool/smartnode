package daemon

import ()


// Config
const checkRPIPVoteInterval string = "10s"


// Run daemon
func Run() {

    // Check for node exit on minipool removed
    go startCheckNodeExit()
    go checkNodeExit()

    // Check RPIP votes periodically
    go startCheckRPIPVotes(checkRPIPVoteInterval)
    go checkRPIPVotes()

    // Block thread
    select {}

}

