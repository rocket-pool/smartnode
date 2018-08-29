package smartnode

import (
    "fmt"
    "time"
)


// Check for node exit on minipool removed
func startCheckNodeExit() {

    // Check for node exit on interval
    // TODO: implement event listener instead
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    for _ = range ticker.C {
        go checkNodeExit()
    }

}


// Check for node exit
func checkNodeExit() {

    // TODO: implement -
    // - check for exit flag; cancel if not set
    //     - get minipool count; cancel if > 0
    //     - exit
    //     - delete exit flag

    // Log
    fmt.Println("Checking for node exit...")

}

