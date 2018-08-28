package daemon

import (
    "fmt"
    "log"
    "time"
)


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

