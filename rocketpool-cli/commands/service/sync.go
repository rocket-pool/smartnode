package service

import (
	"fmt"
	"math"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/terminal"
)

// When printing sync percents, we should avoid printing 100%.
// This function is only called if we're still syncing,
// and the `%0.2f` token will round up if we're above 99.99%.
func SyncRatioToPercent(in float64) float64 {
	return math.Min(99.99, in*100)
	// TODO: INCORPORATE THIS
}

func printClientStatus(status *types.ClientStatus, name string) {

	if status.Error != "" {
		fmt.Printf("Your %s is unavailable (%s).\n", name, status.Error)
		return
	}

	if status.IsSynced {
		fmt.Printf("Your %s is fully synced.\n", name)
		return
	}

	fmt.Printf("Your %s is still syncing (%0.2f%%).\n", name, client.SyncRatioToPercent(status.SyncProgress))
	if strings.Contains(name, "execution") && status.SyncProgress == 0 {
		fmt.Printf("\tNOTE: your %s may not report sync progress.\n\tYou should check its logs to review it.\n", name)
	}
}

func printSyncProgress(status *types.ClientManagerStatus, name string) {

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
	rp := client.NewClientFromCtx(c)

	// Get the config
	cfg, isNew, err := rp.LoadConfig()
	if err != nil {
		return fmt.Errorf("Error loading configuration: %w", err)
	}

	// Print what network we're on
	err = utils.PrintNetwork(cfg.Network.Value, isNew)
	if err != nil {
		return err
	}

	// Make sure ETH2 is on the correct chain
	depositContractInfo, err := rp.Api.Network.GetDepositContractInfo()
	if err != nil {
		return err
	}
	if !depositContractInfo.Data.SufficientSync {
		fmt.Printf("%sYour Execution Client hasn't synced enough to determine if your Execution Client and Beacon Node are on the same network.\n", terminal.ColorYellow)
		fmt.Printf("To run this safety check, try again later when the Execution Client has made more sync progress.%s\n\n", terminal.ColorReset)
	} else if depositContractInfo.Data.RPNetwork != depositContractInfo.Data.BeaconNetwork ||
		depositContractInfo.Data.RPDepositContract != depositContractInfo.Data.BeaconDepositContract {
		utils.PrintDepositMismatchError(
			depositContractInfo.Data.RPNetwork,
			depositContractInfo.Data.BeaconNetwork,
			depositContractInfo.Data.RPDepositContract,
			depositContractInfo.Data.BeaconDepositContract)
		return nil
	} else {
		fmt.Println("Your Beacon Node is on the correct network.")
		fmt.Println()
	}

	// Get node status
	status, err := rp.Api.Service.ClientStatus()
	if err != nil {
		return err
	}

	// Print EC status
	printSyncProgress(&status.Data.EcManagerStatus, "execution")

	// Print CC status
	printSyncProgress(&status.Data.BcManagerStatus, "beacon")

	// Return
	return nil
}
