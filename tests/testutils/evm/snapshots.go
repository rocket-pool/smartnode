package evm

import (
    "github.com/ethereum/go-ethereum/rpc"

    "github.com/rocket-pool/rocketpool-go/tests"
)


// The ID of the current snapshot of the EVM state
var snapshotId string


// Take a snapshot of the EVM state
func TakeSnapshot() error {

    // Initialize RPC client
    client, err := rpc.Dial(tests.Eth1ProviderAddress)
    if err != nil { return err }

    // Make RPC call
    var response string
    if err := client.Call(&response, "evm_snapshot"); err != nil { return err }

    // Set snapshot ID & return
    snapshotId = response
    return nil

}


// Restore a snapshot of the EVM state
func RevertSnapshot() error {

    // Initialize RPC client
    client, err := rpc.Dial(tests.Eth1ProviderAddress)
    if err != nil { return err }

    // Make RPC call & return
    return client.Call(nil, "evm_revert", snapshotId)

}

