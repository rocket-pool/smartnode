package daemons

import (
    "math/big"

    "github.com/ethereum/go-ethereum/core/types"

    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/messaging"
)


// Send block timestamp to a listener channel at interval
func BlockTimeInterval(publisher *messaging.Publisher, interval *big.Int, onInit bool, listener chan *big.Int) {
    blockInterval(publisher, interval, onInit, listener, func(header *types.Header) *big.Int {
        return header.Time
    })
}


// Send block number to a listener channel at interval
func BlockNumberInterval(publisher *messaging.Publisher, interval *big.Int, onInit bool, listener chan *big.Int) {
    blockInterval(publisher, interval, onInit, listener, func(header *types.Header) *big.Int {
        return header.Number
    })
}


// Send to a listener channel at block intervals
func blockInterval(publisher *messaging.Publisher, interval *big.Int, onInit bool, listener chan *big.Int, property func(*types.Header) *big.Int) {

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
            if onInit { listener <- lastCheck }
        } else {
                
            // Get elapsed time since last check
            d := big.NewInt(0)
            d.Sub(property(header), lastCheck)

            // Send if interval has passed
            if d.Cmp(interval) > -1 {
                lastCheck = property(header)
                listener <- lastCheck
            }

        }

    }

}

