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
	if mgrStatus.PrimaryClientStatus.IsSynced {
		rp.SetEcStatusFlags(true, false)
		return nil
	}

	// Fallback EC is good
	if mgrStatus.FallbackEnabled && mgrStatus.FallbackClientStatus.IsSynced {
		if mgrStatus.PrimaryClientStatus.Error != "" {
			fmt.Printf("%sNOTE: primary execution client is unavailable (%s), using fallback execution client...%s\n\n", colorYellow, mgrStatus.PrimaryClientStatus.Error, colorReset)
		} else {
			fmt.Printf("%sNOTE: primary execution client is still syncing (%.2f%%), using fallback execution client...%s\n\n", colorYellow, mgrStatus.PrimaryClientStatus.SyncProgress*100, colorReset)
		}
		rp.SetEcStatusFlags(true, true)
		return nil
	}

	// Is the primary working and syncing?
	if mgrStatus.PrimaryClientStatus.IsWorking && mgrStatus.PrimaryClientStatus.Error == "" {
		if mgrStatus.FallbackEnabled && mgrStatus.FallbackClientStatus.Error != "" {
			return fmt.Errorf("Error: fallback execution client is unavailable (%s), and primary execution client is still syncing (%.2f%%). Please try again later once the client has synced.", mgrStatus.FallbackClientStatus.Error, mgrStatus.PrimaryClientStatus.SyncProgress*100)
		} else {
			return fmt.Errorf("Error: fallback execution client is not configured or unavailable, and primary execution client is still syncing (%.2f%%). Please try again later once the client has synced.", mgrStatus.PrimaryClientStatus.SyncProgress*100)
		}
	}

	// Is the fallback working and syncing?
	if mgrStatus.FallbackEnabled && mgrStatus.FallbackClientStatus.IsWorking && mgrStatus.FallbackClientStatus.Error == "" {
		return fmt.Errorf("Error: primary execution client is unavailable (%s), and fallback execution client is still syncing (%.2f%%). Please try again later.", mgrStatus.PrimaryClientStatus.Error, mgrStatus.FallbackClientStatus.SyncProgress*100)
	}

	// Report if neither client is working
	if mgrStatus.FallbackEnabled {
		return fmt.Errorf("Error: primary execution client is unavailable (%s) and fallback execution client is unavailable (%s), no execution clients are ready.", mgrStatus.PrimaryClientStatus.Error, mgrStatus.FallbackClientStatus.Error)
	} else {
		return fmt.Errorf("Error: primary execution client is unavailable (%s) and no fallback execution client is configured.", mgrStatus.PrimaryClientStatus.Error)
	}

}
