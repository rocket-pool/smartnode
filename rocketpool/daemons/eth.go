package daemons

import (
    "math/big"

    "github.com/ethereum/go-ethereum/core/types"

    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/messaging"
)


// Send to a timer channel at block timestamp intervals
func BlockTimeInterval(publisher *messaging.Publisher, i int64, timer chan bool, onInit bool) {

    // Initialise interval
    interval := big.NewInt(i)

    // Check details
    initialised := false
    lastCheck := big.NewInt(0)

    // Subscribe to new block headers
    newBlockListener := make(chan interface{})
    publisher.AddSubscriber("node.block.new", newBlockListener)

    // Receive headers
    for h := range newBlockListener {
        header := h.(*types.Header)

        // Initialise & send
        if !initialised {
            initialised = true
            lastCheck = header.Time
            if onInit { timer <- true }
        } else {
                
            // Get elapsed time since last check
            d := big.NewInt(0)
            d.Sub(header.Time, lastCheck)

            // Send if interval has passed
            if d.Cmp(interval) > -1 {
                lastCheck = header.Time
                timer <- true
            }

        }

    }

}

