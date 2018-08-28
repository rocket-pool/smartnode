package daemon

import (
    "fmt"
    "log"
    "time"
)


// Check for RPIP alerts on new proposal
func startCheckRPIPAlerts() {

    // Check for alerts on interval
    // TODO: implement event listener instead
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    for _ = range ticker.C {

        // TODO: implement
        // - check for subscription; cancel if not set
        // - get proposal
        // - send alert

        // Log
        fmt.Println("Checking new RPIP for alert...")

    }

}


// Check for RPIP alerts
func checkRPIPAlerts() {

    // TODO: implement
    // - check for subscription; cancel if not set
    // - get new proposals ("ready to commit", not notified)
    // - for each new proposal, send alert

    // Log
    fmt.Println("Checking for RPIP alerts...")

}


// Send RPIP alert
func sendRPIPAlert(proposalId uint64) {

    // TODO: implement
    // - send alert email
    // - mark rpip as notified

}


// Check RPIP votes periodically
func startCheckRPIPVotes(interval string) {

    // Parse check interval
    duration, err := time.ParseDuration(interval)
    if err != nil {
        log.Fatal("Couldn't parse check RPIP votes interval: ", err)
    }

    // Check RPIP votes on interval
    ticker := time.NewTicker(duration)
    defer ticker.Stop()
    for _ = range ticker.C {
        go checkRPIPVotes()
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
    fmt.Println("Checking RPIP votes...")

}

