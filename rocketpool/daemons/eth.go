package daemons

import (
    "math/big"

    "github.com/ethereum/go-ethereum/core/types"

    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/messaging"
)


// Send to a timer channel at block timestamp intervals
func BlockTimeInterval(publisher *messaging.Publisher, i int64, timer chan bool, onInit bool) {
    blockInterval(publisher, i, timer, onInit, func(header *types.Header) *big.Int {
        return header.Time
    })
}


// Send to a timer channel at block number intervals
func BlockNumberInterval(publisher *messaging.Publisher, i int64, timer chan bool, onInit bool) {
    blockInterval(publisher, i, timer, onInit, func(header *types.Header) *big.Int {
        return header.Number
    })
}


// Send to a timer channel at block intervals
func blockInterval(publisher *messaging.Publisher, i int64, timer chan bool, onInit bool, property func(*types.Header) *big.Int) {

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
            lastCheck = property(header)
            if onInit { timer <- true }
        } else {
                
            // Get elapsed time since last check
            d := big.NewInt(0)
            d.Sub(property(header), lastCheck)

            // Send if interval has passed
            if d.Cmp(interval) > -1 {
                lastCheck = property(header)
                timer <- true
            }

        }

    }

}

