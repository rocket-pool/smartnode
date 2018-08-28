package node

import (
    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/storage"
)


// Exit node from network
func ExitNode() (bool, error) {

    // Attempt to exit node
    exited := attemptExitNode()
    if exited {
        return true, nil
    }

    // Open storage
    store, err := storage.Open()
    if err != nil {
        return false, err
    }
    defer store.Close()

    // Store exit flag
    err = store.Put("node.state.exit", true)
    if err != nil {
        return false, err
    }

    // Return
    return false, nil

}


// Attempt to exit node from network
func attemptExitNode() bool {

    // Check for assigned minipools
    // TODO: implement

    // Exit if able
    // TODO: implement

    // Pause
    // TODO: implement
    return false

}

