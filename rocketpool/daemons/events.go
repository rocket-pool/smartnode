package daemons

import (
    "context"
    "log"
    "math/big"

    "github.com/ethereum/go-ethereum"
    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/ethclient"

    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/messaging"
)


// Send events to a listener channel on found
func SendEvents(publisher *messaging.Publisher, client *ethclient.Client, contractAddress *common.Address, contractAbi *abi.ABI, eventName string, eventPrototype interface{}, listener chan *interface{}) {

    // Create contract instance
    contract := bind.NewBoundContract(*contractAddress, *contractAbi, client, client, client)

    // Subscribe to new block headers
    newBlockListener := make(chan interface{})
    publisher.AddSubscriber("node.block.new", newBlockListener)

    // Receive headers
    for h := range newBlockListener {
        header := h.(*types.Header)

        // Get logs in block by contract address & event topic
        logs, err := client.FilterLogs(context.Background(), ethereum.FilterQuery{
            FromBlock: header.Number,
            ToBlock: header.Number,
            Addresses: []common.Address{*contractAddress},
            Topics: [][]common.Hash{{contractAbi.Events[eventName].Id()}},
        })
        if err != nil {
            log.Fatal("Error retrieving logs: ", err)
        }

        // Process logs
        for _, logItem := range logs {

            // Unpack event data from log
            event := eventPrototype
            err := contract.UnpackLog(event, eventName, logItem)
            if err != nil {
                log.Fatal("Error unpacking event from log: ", err)
            }

            // Send event to listener
            listener <- &event

        }

    }

}


// Send block timestamp to a listener channel at interval
func SendBlockTimeIntervals(publisher *messaging.Publisher, interval *big.Int, onInit bool, listener chan *big.Int) {
    sendBlockIntervals(publisher, interval, onInit, listener, func(header *types.Header) *big.Int {
        return header.Time
    })
}


// Send block number to a listener channel at interval
func SendBlockNumberIntervals(publisher *messaging.Publisher, interval *big.Int, onInit bool, listener chan *big.Int) {
    sendBlockIntervals(publisher, interval, onInit, listener, func(header *types.Header) *big.Int {
        return header.Number
    })
}


// Send to a listener channel at block intervals
func sendBlockIntervals(publisher *messaging.Publisher, interval *big.Int, onInit bool, listener chan *big.Int, property func(*types.Header) *big.Int) {

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

