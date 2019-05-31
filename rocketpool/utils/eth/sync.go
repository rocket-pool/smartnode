package eth

import (
    "context"
    "errors"
    "fmt"

    "github.com/ethereum/go-ethereum/ethclient"
)


// Wait for node to sync
func WaitSync(client *ethclient.Client, renderStatus bool) error {

    // Status channels
    successChannel := make(chan bool)
    errorChannel := make(chan error)

    // Check node sync status
    go (func() {
        var statusRendered bool = false
        var checkSync bool = true
        for checkSync {

            // Check sync progress and render
            if progress, err := client.SyncProgress(context.Background()); err != nil {
                checkSync = false
                if statusRendered { fmt.Println("") }
                errorChannel <- errors.New("Error retrieving ethereum node sync progress: " + err.Error())
            } else if progress == nil {
                checkSync = false
                if statusRendered { fmt.Println("") }
                successChannel <- true
            } else if renderStatus {
                if statusRendered { fmt.Print("\r") }
                fmt.Printf("Node syncing: %.2f%% ", (float64(progress.CurrentBlock - progress.StartingBlock) / float64(progress.HighestBlock - progress.StartingBlock)) * 100)
                statusRendered = true
            }

        }
    })()

    // Receive status and return
    select {
        case <-successChannel:
            return nil
        case err := <-errorChannel:
            return err
    }

}

