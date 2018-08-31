package rpip

import (
    "math/big"

    "github.com/rocket-pool/smartnode-cli/rocketpool/daemons"
    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/messaging"
)


// Check for RPIP alerts on new proposal
func StartCheckRPIPAlerts() {

    // TODO: implement
    // - check for subscription; cancel if not set
    // - get proposal
    // - send alert

}


// Check RPIP votes on block timestamp
func StartCheckRPIPVotes(publisher *messaging.Publisher, interval int64) {

    // Create block timestamp interval listener
    listener := make(chan *big.Int)
    go daemons.SendBlockTimeIntervals(publisher, big.NewInt(interval), true, listener)

    // Check RPIP votes on interval
    for _ = range listener {
        checkRPIPVotes()
    }

}

