package lighthouse

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/wallet/keystore"
)

// Config
const (
	FileMode fs.FileMode = 0600
	DirMode  fs.FileMode = 0700
)

type FeeRecipientManager struct {
	keystore keystore.Keystore
}

// Creates a new fee recipient manager
func NewFeeRecipientManager(keystore keystore.Keystore) *FeeRecipientManager {
	return &FeeRecipientManager{
		keystore: keystore,
	}
}

// Checks if the fee recipient file exists and has the correct distributor address in it.
// The first return value is for file existence, the second is for validation of the fee recipient address inside.
func (fm *FeeRecipientManager) CheckFeeRecipientFile(distributor common.Address) (bool, bool, error) {

	// Check if the file exists
	path := filepath.Join(fm.keystore.GetKeystoreDir(), config.LighthouseFeeRecipientFilename)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, false, nil
	} else if err != nil {
		return false, false, err
	}

	// Create the distributor address string for the node
	distributorAddress := distributor.Hex()
	expectedString := distributorAddress

	// Compare the file contents with the expected string
	bytes, err := ioutil.ReadFile(path)
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
func (fm *FeeRecipientManager) UpdateFeeRecipientFile(distributor common.Address) error {

	// Create the distributor address string for the node
	distributorAddress := distributor.Hex()
	expectedString := distributorAddress
	bytes := []byte(expectedString)

	// Create keystore dir
	if err := os.MkdirAll(fm.keystore.GetKeystoreDir(), DirMode); err != nil {
		return fmt.Errorf("Could not create fee recipient folder [%s]: %w", fm.keystore.GetKeystoreDir(), err)
	}

	// Write the file
	path := filepath.Join(fm.keystore.GetKeystoreDir(), config.LighthouseFeeRecipientFilename)
	err := ioutil.WriteFile(path, bytes, FileMode)
	if err != nil {
		return fmt.Errorf("error writing fee recipient file: %w", err)
	}
	return nil

}
