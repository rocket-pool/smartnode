package node

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


func getSyncProgress(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Make sure ETH2 is on the correct chain
    depositContractInfo, err := rp.DepositContractInfo()
    if err != nil {
        return err
    }
    if depositContractInfo.RPNetwork != depositContractInfo.BeaconNetwork ||
       depositContractInfo.RPDepositContract != depositContractInfo.BeaconDepositContract {
        cliutils.PrintDepositMismatchError(
            depositContractInfo.RPNetwork,
            depositContractInfo.BeaconNetwork,
            depositContractInfo.RPDepositContract,
            depositContractInfo.BeaconDepositContract)
        return nil
    } else {
        fmt.Println("Your eth2 client is on the correct network.\n")
    }

    // Get node status
    status, err := rp.NodeSync()
    if err != nil {
        return err
    }

    // Print eth1 status
    if status.Eth1Synced {
        fmt.Print("Your eth1 client is fully synced.\n")
    } else {
        fmt.Printf("Your eth1 client is still syncing (%0.2f%%).\n", status.Eth1Progress * 100)
    }

    // Print eth2 status
    if status.Eth2Synced {
        fmt.Print("Your eth2 client is fully synced.\n")
    } else if status.Eth2Progress != -1 {
        fmt.Printf("Your eth2 client is still syncing (%0.2f%%).\n", status.Eth2Progress * 100)
    } else {
        fmt.Print("Your eth2 client is still syncing (but does not provide its progress).\n")
    }

    // Return
    return nil

}

