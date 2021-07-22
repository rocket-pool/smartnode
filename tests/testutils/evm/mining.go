package evm

import (
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/rocket-pool/rocketpool-go/tests"
)

// Mine a number of blocks
func MineBlocks(numBlocks int) error {

    // Initialize RPC client
    client, err := rpc.Dial(tests.Eth1ProviderAddress)
    if err != nil { return err }

    // Make RPC calls
    for bi := 0; bi < numBlocks; bi++ {
        if err := client.Call(nil, "evm_mine"); err != nil { return err }
    }

    // Return
    return nil

}


// Fast forward to some number of seconds
func IncreaseTime(time int) error {

    // Initialize RPC client
    client, err := rpc.Dial(tests.Eth1ProviderAddress)
    if err != nil { return err }

    // Make RPC calls
    if err := client.Call(nil, "evm_increaseTime", time); err != nil { return err }
    if err := MineBlocks(1); err != nil { return err }

    // Return
    return nil

}

