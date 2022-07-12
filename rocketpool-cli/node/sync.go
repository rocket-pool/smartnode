package node

import (
	"fmt"
	"time"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// Settings
var ethClientRecentBlockThreshold, _ = time.ParseDuration("5m")

func getSyncProgress(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c)
	if err != nil {
		return err
	}
	defer rp.Close()

	// Print what network we're on
	err = cliutils.PrintNetwork(rp)
	if err != nil {
		return err
	}

	// Make sure ETH2 is on the correct chain
	depositContractInfo, err := rp.DepositContractInfo()
	if err != nil {
		return err
	}
	if !depositContractInfo.SufficientSync {
		colorReset := "\033[0m"
		colorYellow := "\033[33m"
		fmt.Printf("%sYour eth1 client hasn't synced enough to determine if your eth1 and eth2 clients are on the same network.\n", colorYellow)
		fmt.Printf("To run this safety check, try again later when eth1 has made more sync progress.%s\n\n", colorReset)
	} else if depositContractInfo.RPNetwork != depositContractInfo.BeaconNetwork ||
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

	// Print EC status
	if status.EcStatus.PrimaryClientStatus.Error != "" {
		fmt.Printf("Your primary execution client is unavailable (%s).\n", status.EcStatus.PrimaryClientStatus.Error)
	} else if status.EcStatus.PrimaryClientStatus.IsSynced {
		fmt.Print("Your primary execution client is fully synced.\n")
	} else {
		fmt.Printf("Your primary execution client is still syncing (%0.2f%%).\n", status.EcStatus.PrimaryClientStatus.SyncProgress*100)
		if status.EcStatus.PrimaryClientStatus.SyncProgress == 0 {
			fmt.Println("\tNOTE: your execution client may not report sync progress.\n\tYou should check your its logs to review it.")
		}
	}

	// Print fallback EC status
	if status.EcStatus.FallbackEnabled {
		if status.EcStatus.FallbackClientStatus.Error != "" {
			fmt.Printf("Your fallback execution client is unavailable (%s).\n", status.EcStatus.FallbackClientStatus.Error)
		} else if status.EcStatus.FallbackClientStatus.IsSynced {
			fmt.Print("Your fallback execution client is fully synced.\n")
		} else {
			fmt.Printf("Your fallback execution client is still syncing (%0.2f%%).\n", status.EcStatus.FallbackClientStatus.SyncProgress*100)
			if status.EcStatus.FallbackClientStatus.SyncProgress == 0 {
				fmt.Println("\tNOTE: your execution client may not report sync progress.\n\tYou should check your its logs to review it.")
			}
		}
	} else {
		fmt.Printf("You do not have a fallback execution client enabled.\n")
	}

	// Print CC status
	if status.BcStatus.PrimaryClientStatus.Error != "" {
		fmt.Printf("Your primary consensus client is unavailable (%s).\n", status.BcStatus.PrimaryClientStatus.Error)
	} else if status.BcStatus.PrimaryClientStatus.IsSynced {
		fmt.Print("Your primary consensus client is fully synced.\n")
	} else {
		fmt.Printf("Your primary consensus client is still syncing (%0.2f%%).\n", status.BcStatus.PrimaryClientStatus.SyncProgress*100)
	}

	// Print fallback CC status
	if status.BcStatus.FallbackEnabled {
		if status.BcStatus.FallbackClientStatus.Error != "" {
			fmt.Printf("Your fallback consensus client is unavailable (%s).\n", status.BcStatus.FallbackClientStatus.Error)
		} else if status.BcStatus.FallbackClientStatus.IsSynced {
			fmt.Print("Your fallback consensus client is fully synced.\n")
		} else {
			fmt.Printf("Your fallback consensus client is still syncing (%0.2f%%).\n", status.BcStatus.FallbackClientStatus.SyncProgress*100)
		}
	} else {
		fmt.Printf("You do not have a fallback consensus client enabled.\n")
	}

	// Return
	return nil

}
