package node

import (
	"fmt"
	"strings"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

func printClientStatus(status *api.ClientStatus, name string) {

	if status.Error != "" {
		fmt.Printf("Your %s is unavailable (%s).\n", name, status.Error)
		return
	}

	if status.IsSynced {
		fmt.Printf("Your %s is fully synced.\n", name)
		return
	}

	fmt.Printf("Your %s is still syncing (%0.2f%%).\n", name, rocketpool.SyncRatioToPercent(status.SyncProgress))
	if strings.Contains(name, "execution") && status.SyncProgress == 0 {
		fmt.Printf("\tNOTE: your %s may not report sync progress.\n\tYou should check its logs to review it.\n", name)
	}
}

func printSyncProgress(status *api.ClientManagerStatus, name string) {

	// Print primary client status
	printClientStatus(&status.PrimaryClientStatus, fmt.Sprintf("primary %s client", name))

	if !status.FallbackEnabled {
		fmt.Printf("You do not have a fallback %s client enabled.\n", name)
		return
	}

	// A fallback is enabled, so print fallback client status
	printClientStatus(&status.FallbackClientStatus, fmt.Sprintf("fallback %s client", name))
}

func getSyncProgress(c *cli.Context) error {

	// Get RP client
	rp := rocketpool.NewClientFromCtx(c)
	defer rp.Close()

	// Get the config
	cfg, isNew, err := rp.LoadConfig()
	if err != nil {
		return fmt.Errorf("Error loading configuration: %w", err)
	}

	// Print what network we're on
	err = cliutils.PrintNetwork(cfg.GetNetwork(), isNew)
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
		fmt.Printf("%sYour execution client hasn't synced enough to determine if your execution and consensus clients are on the same network.\n", colorYellow)
		fmt.Printf("To run this safety check, try again later when the execution client has made more sync progress.%s\n\n", colorReset)
	} else if depositContractInfo.RPNetwork != depositContractInfo.BeaconNetwork ||
		depositContractInfo.RPDepositContract != depositContractInfo.BeaconDepositContract {
		cliutils.PrintDepositMismatchError(
			depositContractInfo.RPNetwork,
			depositContractInfo.BeaconNetwork,
			depositContractInfo.RPDepositContract,
			depositContractInfo.BeaconDepositContract)
		return nil
	} else {
		fmt.Println("Your consensus client is on the correct network.\n")
	}

	// Get node status
	status, err := rp.NodeSync()
	if err != nil {
		return err
	}

	// Print EC status
	printSyncProgress(&status.EcStatus, "execution")

	// Print CC status
	printSyncProgress(&status.BcStatus, "consensus")

	// Return
	return nil

}
