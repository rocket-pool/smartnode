package cli

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
)

// Check the status of the execution client(s) and provision the API with them
func CheckExecutionClientStatus(rp *rocketpool.Client) error {

	// Check if the primary EC is up, synced, and able to respond to requests - if not, forces the use of the fallback EC for this command
	response, err := rp.GetExecutionClientStatus()
	if err != nil {
		return err
	}

	mgrStatus := response.ManagerStatus

	// Primary EC is good
	if mgrStatus.PrimaryEcStatus.IsSynced {
		rp.SetEcStatusFlags(true, false)
	}

	// Fallback EC is good
	if mgrStatus.FallbackEnabled && mgrStatus.FallbackEcStatus.IsSynced {
		if mgrStatus.PrimaryEcStatus.Error != "" {
			fmt.Printf("%sNOTE: primary execution client is unavailable (%s), using fallback execution client...%s\n\n", colorYellow, mgrStatus.PrimaryEcStatus.Error, colorReset)
		} else {
			fmt.Printf("%sNOTE: primary execution client is still syncing (%.2f%%), using fallback execution client...%s\n\n", colorYellow, mgrStatus.PrimaryEcStatus.SyncProgress*100, colorReset)
		}
		rp.SetEcStatusFlags(true, true)
		return nil
	}

	// Is the primary working and syncing?
	if mgrStatus.PrimaryEcStatus.IsWorking && mgrStatus.PrimaryEcStatus.Error == "" {
		if mgrStatus.FallbackEnabled && mgrStatus.FallbackEcStatus.Error != "" {
			return fmt.Errorf("Error: fallback execution client is unavailable (%s), and primary execution client is still syncing (%.2f%%). Please try again later once the client has synced.", mgrStatus.FallbackEcStatus.Error, mgrStatus.PrimaryEcStatus.SyncProgress*100)
		} else {
			return fmt.Errorf("Error: fallback execution client is not configured or unavailable, and primary execution client is still syncing (%.2f%%). Please try again later once the client has synced.", mgrStatus.PrimaryEcStatus.SyncProgress*100)
		}
	}

	// Is the fallback working and syncing?
	if mgrStatus.FallbackEnabled && mgrStatus.FallbackEcStatus.IsWorking && mgrStatus.FallbackEcStatus.Error == "" {
		return fmt.Errorf("Error: primary execution client is unavailable (%s), and fallback execution client is still syncing (%.2f%%). Please try again later.", mgrStatus.PrimaryEcStatus.Error, mgrStatus.FallbackEcStatus.SyncProgress*100)
	}

	// Report if neither client is working
	if mgrStatus.FallbackEnabled {
		return fmt.Errorf("Error: primary execution client is unavailable (%s) and fallback execution client is unavailable (%s), no execution clients are ready.", mgrStatus.PrimaryEcStatus.Error, mgrStatus.FallbackEcStatus.Error)
	} else {
		return fmt.Errorf("Error: primary execution client is unavailable (%s) and no fallback execution client is configured.", mgrStatus.PrimaryEcStatus.Error)
	}

}
