package smartnode

import (
    "log"

    "github.com/rocket-pool/smartnode-cli/rocketpool/daemons"
    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/messaging"
)


// Check RPIP votes on block timestamp
func startCheckRPIPVotes(publisher *messaging.Publisher, interval int64) {

    // Create block timestamp interval timer
    timer := make(chan bool)
    go daemons.BlockTimeInterval(publisher, interval, timer, true)

    // Check RPIP votes on interval
    for _ = range timer {
        checkRPIPVotes()
    }

}


// Check RPIP votes
func checkRPIPVotes() {

    // TODO: implement -
    // - get "ready to reveal" proposals
    // - for each "ready to reveal" proposal:
    //     - check for stored vote; cancel if not set
    //     - reveal vote
    //     - delete stored vote

    // Log
    log.Println("Checking RPIP votes...")

}

