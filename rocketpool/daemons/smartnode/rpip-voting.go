package smartnode

import (
    "log"
    "math/big"

    "github.com/rocket-pool/smartnode-cli/rocketpool/daemons"
    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/messaging"
)


// Check RPIP votes on block timestamp
func startCheckRPIPVotes(publisher *messaging.Publisher, interval int64) {

    // Create block timestamp interval listener
    listener := make(chan *big.Int)
    go daemons.SendBlockTimeIntervals(publisher, big.NewInt(interval), true, listener)

    // Check RPIP votes on interval
    for _ = range listener {
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

