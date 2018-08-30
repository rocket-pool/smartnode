package daemons

import (
    "context"
    "log"
    "math/big"

    "github.com/ethereum/go-ethereum"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/ethclient"

    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/eth"
    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/messaging"
)


// Send logs in block by topic to a listener channel on found
func BlockLog(publisher *messaging.Publisher, client *ethclient.Client, topic string, listener chan *types.Log) {

    // Subscribe to new block headers
    newBlockListener := make(chan interface{})
    publisher.AddSubscriber("node.block.new", newBlockListener)

    // Receive headers
    for h := range newBlockListener {
        header := h.(*types.Header)

        // Get logs in block by topic
        logs, err := client.FilterLogs(context.Background(), ethereum.FilterQuery{
            FromBlock: header.Number,
            ToBlock: header.Number,
            Topics: [][]common.Hash{{eth.KeccakStr(topic)}},
        })
        if err != nil {
            log.Fatal("Error retrieving logs: ", err)
        }

        // Send logs
        for _, logItem := range logs {
            listener <- &logItem
        }

    }

}


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

