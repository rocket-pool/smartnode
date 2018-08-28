package daemon

import (
    "fmt"
    "log"
    "time"
)


// Config
const checkRPIPInterval string = "10s"


// Run daemon
func Run() {

    // Check RPIPs periodically
    go checkRPIPs(checkRPIPInterval)

    // Block thread
    select {}

}


// Check RPIPs periodically
func checkRPIPs(interval string) {

    // Parse check interval
    duration, err := time.ParseDuration(interval)
    if err != nil {
        log.Fatal("Couldn't parse check RPIPs interval: ", err)
    }

    // Check RPIPs on interval
    ticker := time.NewTicker(duration)
    defer ticker.Stop()
    for _ = range ticker.C {
        fmt.Println("Checking RPIPs...")
    }

}

