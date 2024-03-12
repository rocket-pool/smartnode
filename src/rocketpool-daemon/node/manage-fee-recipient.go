package node

import (
	"fmt"
	"io/fs"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/beacon"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/log"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/state"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/validator"
	"github.com/rocket-pool/smartnode/shared/config"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Config
const (
	FileMode fs.FileMode = 0644
)

// Manage fee recipient task
type ManageFeeRecipient struct {
	sp  *services.ServiceProvider
	cfg *config.RocketPoolConfig
	log log.ColorLogger
	bc  beacon.Client
}

// Create manage fee recipient task
func NewManageFeeRecipient(sp *services.ServiceProvider, logger log.ColorLogger) *ManageFeeRecipient {
	return &ManageFeeRecipient{
		sp:  sp,
		log: logger,
	}
}

// Manage fee recipient
func (t *ManageFeeRecipient) Run(state *state.NetworkState) error {
	// Get services
	t.cfg = t.sp.GetConfig()
	t.bc = t.sp.GetBeaconClient()
	d := t.sp.GetDocker()
	nodeAddress, _ := t.sp.GetWallet().GetAddress()

	// Log
	t.log.Println("Checking for correct fee recipient...")

	// Get the fee recipient info for the node
	feeRecipientInfo, err := t.getFeeRecipientInfo(nodeAddress, state)
	if err != nil {
		return fmt.Errorf("error getting fee recipient info: %w", err)
	}

	// Get the correct fee recipient address
	var correctFeeRecipient common.Address
	if feeRecipientInfo.IsInSmoothingPool || feeRecipientInfo.IsInOptOutCooldown {
		correctFeeRecipient = feeRecipientInfo.SmoothingPoolAddress
	} else {
		correctFeeRecipient = feeRecipientInfo.FeeDistributorAddress
	}

	// Check if the VC is using the correct fee recipient
	fileExists, correctAddress, err := t.checkFeeRecipientFile(correctFeeRecipient)
	if err != nil {
		return fmt.Errorf("error validating fee recipient files: %w", err)
	}

	if !fileExists {
		t.log.Println("Fee recipient files don't all exist, regenerating...")
	} else if !correctAddress {
		t.log.Printlnf("WARNING: Fee recipient files did not contain the correct fee recipient of %s, regenerating...", correctFeeRecipient.Hex())
	} else {
		// Files are all correct, return.
		return nil
	}

	// Regenerate the fee recipient files
	err = t.updateFeeRecipientFile(correctFeeRecipient)
	if err != nil {
		t.log.Println("***ERROR***")
		t.log.Printlnf("Error updating fee recipient files: %s", err.Error())
		t.log.Println("Shutting down the validator client for safety to prevent you from being penalized...")

		err = validator.StopValidator(t.cfg, t.bc, &t.log, d)
		if err != nil {
			return fmt.Errorf("error stopping validator client: %w", err)
		}
		return nil
	}

	// Restart the VC
	t.log.Println("Fee recipient files updated successfully! Restarting validator client...")
	err = validator.RestartValidator(t.cfg, t.bc, &t.log, d)
	if err != nil {
		return fmt.Errorf("error restarting validator client: %w", err)
	}

	// Log & return
	t.log.Println("Successfully restarted, you are now validating safely.")
	return nil
}

// Get info about the node's fee recipient
func (t *ManageFeeRecipient) getFeeRecipientInfo(nodeAddress common.Address, state *state.NetworkState) (*api.FeeRecipientInfo, error) {
	info := &api.FeeRecipientInfo{
		IsInOptOutCooldown: false,
		OptOutEpoch:        0,
	}

	mpd := state.NodeDetailsByAddress[nodeAddress]

	// Get info
	info.SmoothingPoolAddress = state.NetworkDetails.SmoothingPoolAddress
	info.FeeDistributorAddress = mpd.FeeDistributorAddress
	info.IsInSmoothingPool = mpd.SmoothingPoolRegistrationState

	// Calculate the safe opt-out epoch if applicable
	if !info.IsInSmoothingPool {
		// Get the opt out time
		optOutTime := time.Unix(mpd.SmoothingPoolRegistrationChanged.Int64(), 0)

		// Get the Beacon info
		beaconConfig := state.BeaconConfig
		beaconHead, err := t.bc.GetBeaconHead()
		if err != nil {
			return nil, fmt.Errorf("Error getting Beacon head: %w", err)
		}

		// Check if the user just opted out
		if optOutTime != time.Unix(0, 0) {
			// Get the epoch for that time
			genesisTime := time.Unix(int64(beaconConfig.GenesisTime), 0)
			secondsSinceGenesis := optOutTime.Sub(genesisTime)
			epoch := uint64(secondsSinceGenesis.Seconds()) / beaconConfig.SecondsPerEpoch

			// Make sure epoch + 1 is finalized - if not, they're still on cooldown
			targetEpoch := epoch + 1
			if beaconHead.FinalizedEpoch < targetEpoch {
				info.IsInOptOutCooldown = true
				info.OptOutEpoch = targetEpoch
			}
		}
	}

	return info, nil
}

// Checks if the fee recipient file exists and has the correct distributor address in it.
// The first return value is for file existence, the second is for validation of the fee recipient address inside.
func (t *ManageFeeRecipient) checkFeeRecipientFile(feeRecipient common.Address) (bool, bool, error) {
	// Check if the file exists
	path := t.cfg.Smartnode.GetFeeRecipientFilePath()
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, false, nil
	} else if err != nil {
		return false, false, err
	}

	// Compare the file contents with the expected string
	expectedString := t.getFeeRecipientFileContents(feeRecipient)
	bytes, err := os.ReadFile(path)
	if err != nil {
		return false, false, fmt.Errorf("error reading fee recipient file: %w", err)
	}
	existingString := string(bytes)
	if existingString != expectedString {
		// If it wrote properly, indicate a success but that the file needed to be updated
		return true, false, nil
	}

	// The file existed and had the expected address, all set.
	return true, true, nil
}

// Writes the given address to the fee recipient file. The VC should be restarted to pick up the new file.
func (t *ManageFeeRecipient) updateFeeRecipientFile(feeRecipient common.Address) error {

	// Create the distributor address string for the node
	expectedString := t.getFeeRecipientFileContents(feeRecipient)
	bytes := []byte(expectedString)

	// Write the file
	path := t.cfg.Smartnode.GetFeeRecipientFilePath()
	err := os.WriteFile(path, bytes, FileMode)
	if err != nil {
		return fmt.Errorf("error writing fee recipient file: %w", err)
	}
	return nil

}

// Gets the expected contents of the fee recipient file
func (t *ManageFeeRecipient) getFeeRecipientFileContents(feeRecipient common.Address) string {
	if !t.cfg.IsNativeMode {
		// Docker mode
		return feeRecipient.Hex()
	}

	// Native mode
	return fmt.Sprintf("FEE_RECIPIENT=%s", feeRecipient.Hex())
}
