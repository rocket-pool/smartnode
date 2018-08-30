package smartnode

import (
    "log"
)


// Check for RPIP alerts on new proposal
func startCheckRPIPAlerts() {

    // TODO: implement
    // - check for subscription; cancel if not set
    // - get proposal
    // - send alert

}


// Check for RPIP alerts
func checkRPIPAlerts() {

    // TODO: implement
    // - check for subscription; cancel if not set
    // - get new proposals ("ready to commit", not notified)
    // - for each new proposal, send alert

    // Log
    log.Println("Checking for RPIP alerts...")

}


// Send RPIP alert
func sendRPIPAlert(proposalId uint64) {

    // TODO: implement
    // - send alert email
    // - mark rpip as notified

}

