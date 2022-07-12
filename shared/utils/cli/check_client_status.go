package cli

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Check the status of the Execution and Consensus client(s) and provision the API with them
func CheckClientStatus(rp *rocketpool.Client) error {

	// Check if the primary clients are up, synced, and able to respond to requests - if not, forces the use of the fallbacks for this command
	response, err := rp.GetClientStatus()
	if err != nil {
		return err
	}

	ecMgrStatus := response.EcManagerStatus
	bcMgrStatus := response.BcManagerStatus

	// Primary EC and CC are good
	if ecMgrStatus.PrimaryClientStatus.IsSynced && bcMgrStatus.PrimaryClientStatus.IsSynced {
		rp.SetClientStatusFlags(true, false)
		return nil
	}

	// Get the status messages
	primaryEcStatus := getClientStatusString(ecMgrStatus.PrimaryClientStatus)
	primaryBcStatus := getClientStatusString(bcMgrStatus.PrimaryClientStatus)
	fallbackEcStatus := getClientStatusString(ecMgrStatus.FallbackClientStatus)
	fallbackBcStatus := getClientStatusString(bcMgrStatus.FallbackClientStatus)

	// Check the fallbacks if enabled
	if ecMgrStatus.FallbackEnabled && bcMgrStatus.FallbackEnabled {

		// Fallback EC and CC are good
		if ecMgrStatus.FallbackClientStatus.IsSynced && bcMgrStatus.FallbackClientStatus.IsSynced {
			fmt.Printf("%sNOTE: primary clients are not ready, using fallback clients...\n\tPrimary EC status: %s\n\tPrimary CC status: %s%s\n\n", colorYellow, primaryEcStatus, primaryBcStatus, colorReset)
			rp.SetClientStatusFlags(true, true)
			return nil
		}

		// Both pairs aren't ready
		return fmt.Errorf("Error: neither primary nor fallback client pairs are ready.\n\tPrimary EC status: %s\n\tPrimary CC status: %s\n\tFallback EC status: %s\n\tFallback CC status: %s", primaryEcStatus, primaryBcStatus, fallbackEcStatus, fallbackBcStatus)

	}

	// Primary isn't ready and fallback isn't enabled
	return fmt.Errorf("Error: primary client pair isn't ready and fallback clients aren't enabled.\n\tPrimary EC status: %s\n\tPrimary CC status: %s", primaryEcStatus, primaryBcStatus)

}

func getClientStatusString(clientStatus api.ClientStatus) string {
	if clientStatus.IsSynced {
		return "synced and ready"
	} else if clientStatus.IsWorking {
		return fmt.Sprintf("syncing (%.2f%%)", clientStatus.SyncProgress)
	} else {
		return fmt.Sprintf("unavailable (%s)", clientStatus.Error)
	}
}
