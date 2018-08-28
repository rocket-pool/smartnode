package daemon

import ()


// Config
const checkRPIPVoteInterval string = "10s"


// Run daemon
func Run() {

    // Check RPIP votes periodically
    go startCheckRPIPVotes(checkRPIPVoteInterval)
    go checkRPIPVotes()

    // Block thread
    select {}

}

