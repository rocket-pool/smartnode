package smartnode

import (
    "context"
    "log"

    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/ethereum/go-ethereum/core/types"

    "github.com/rocket-pool/smartnode-cli/rocketpool/daemons/smartnode/modules/rpip"
    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/messaging"
)


// Config
const clientUrl string = "ws://localhost:8545" // Ganache
const checkRPIPVoteInterval int64 = 10 // Seconds


// Run daemon
func Run() {

    // Initilise pubsub
    publisher := messaging.NewPublisher()

    // Connect to ethereum node
    client, err := ethclient.Dial(clientUrl)
    if err != nil {
        log.Fatal("Error connecting to ethereum node: ", err)
    }

    // Listen for new block headers and notify
    go (func() {

        // Subscribe to new headers
        headers := make(chan *types.Header)
        sub, err := client.SubscribeNewHead(context.Background(), headers)
        if err != nil {
            log.Fatal("Error subscribing to block headers: ", err)
        }

        // Receive headers and notify
        for {
            select {
                case err := <-sub.Err():
                    log.Fatal("Block header subscription error: ", err)
                case header := <-headers:
                    publisher.Notify("node.block.new", header)
            }
        }

    })()

    // Check RPIP votes on block timestamp
    go rpip.StartCheckRPIPVotes(publisher, checkRPIPVoteInterval)

    // Block thread
    select {}

}

