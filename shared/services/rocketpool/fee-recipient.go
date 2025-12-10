package rocketpool

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/shared/services/config"
)

// Config
const (
	FileMode fs.FileMode = 0644
)

// Checks if the fee recipient file exists and has the correct distributor address in it.
// The first return value is for file existence, the second is for validation of the fee recipient address inside.
func CheckFeeRecipientFile(feeRecipient common.Address, cfg *config.RocketPoolConfig) (bool, bool, error) {

	// Check if the file exists
	path := cfg.Smartnode.GetGlobalFeeRecipientFilePath()
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, false, nil
	} else if err != nil {
		return false, false, err
	}

	// Compare the file contents with the expected string
	expectedString := getGlobalFeeRecipientFileContents(feeRecipient, cfg)
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

// Writes the given address to the globalfee recipient file. The VC should be restarted to pick up the new file.
func UpdateGlobalFeeRecipientFile(feeRecipient common.Address, cfg *config.RocketPoolConfig) error {

	// Create the distributor address string for the node
	expectedString := getGlobalFeeRecipientFileContents(feeRecipient, cfg)
	bytes := []byte(expectedString)

	// Write the file
	globalPath := cfg.Smartnode.GetGlobalFeeRecipientFilePath()
	err := os.WriteFile(globalPath, bytes, FileMode)
	if err != nil {
		return fmt.Errorf("error writing global fee recipient file: %w", err)
	}
	return nil

}

// Writes the given address to the per key fee recipient file. The VC should be restarted to pick up the new file.
func UpdatePerKeyFeeRecipientFiles(pubkeys []types.ValidatorPubkey, megapoolAddress common.Address, cfg *config.RocketPoolConfig) error {
	// Check which beacon client is being used
	// cc, mode := cfg.GetSelectedConsensusClient()
	// path := cfg.Smartnode.GetPerKeyFeeRecipientFilePath() + "-" + string(cc)

	// switch cc {
	// case cfgtypes.ConsensusClient_Lighthouse:
	// 	path = cfg.Smartnode.GetGlobalFeeRecipientFilePath()
	// case cfgtypes.ConsensusClient_Lodestar:
	// 	path = cfg.Smartnode.GetGlobalFeeRecipientFilePath()
	// case cfgtypes.ConsensusClient_Nimbus:
	// 	path = cfg.Smartnode.GetGlobalFeeRecipientFilePath()
	// case cfgtypes.ConsensusClient_Prysm:
	// 	path = cfg.Smartnode.GetGlobalFeeRecipientFilePath()
	// case cfgtypes.ConsensusClient_Teku:

	// }
	// // Create the per key fee recipient files
	// for _, pubkey := range pubkeys {

	// }
	return nil
}

// Gets the expected contents of the fee recipient file
func getGlobalFeeRecipientFileContents(feeRecipient common.Address, cfg *config.RocketPoolConfig) string {
	if !cfg.IsNativeMode {
		// Docker mode
		return feeRecipient.Hex()
	}

	// Native mode
	return fmt.Sprintf("FEE_RECIPIENT=%s", feeRecipient.Hex())
}

func getLighthousePerKeyFeeRecipientFileContents(pubkeys []types.ValidatorPubkey, feeRecipient common.Address) string {
	if len(pubkeys) == 0 {
		return ""
	}

	// Iterate pubkeys to create a json file

	return ""
}
