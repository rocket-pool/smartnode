package node

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
)


func getSyncProgress(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Get node status
    status, err := rp.NodeSync()
    if err != nil {
        return err
    }

    // Print eth1 status
    if status.Eth1Synced {
        fmt.Print("Your eth1 client is fully synced.\n")
    } else {
        fmt.Printf("Your eth1 client is still syncing (%0.2f%%)\n", status.Eth1Progress * 100)
    }

    // Print eth2 status
    if status.Eth2Synced {
        fmt.Print("Your eth2 client is fully synced.\n")
    } else if status.Eth2Progress != -1 {
        fmt.Printf("Your eth2 client is still syncing (%0.2f%%)\n", status.Eth2Progress * 100)
    } else {
        fmt.Print("Your eth2 client is still syncing (but does not provide its progress).\n")
    }

    // Return
    return nil

}

