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
// If it does, this returns true - the file is up to date.
// Otherwise, this writes the file and returns false indicating that the VC should be restarted to pick up the new file.
func (fm *FeeRecipientManager) CheckAndUpdateFeeRecipientFile(distributor common.Address) (bool, error) {

	// Create the distributor address string for the node
	distributorAddress := distributor.Hex()
	expectedString := fmt.Sprintf("default: %s\n", distributorAddress)

	// Check if the file exists, and write it if it doesn't
	path := filepath.Join(fm.keystore.GetKeystoreDir(), config.LighthouseFeeRecipientFilename)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		bytes := []byte(expectedString)
		err = ioutil.WriteFile(path, bytes, FileMode)
		if err != nil {
			return false, fmt.Errorf("error writing fee recipient file: %w", err)
		}

		// If it wrote properly, indicate a success but that the file needed to be updated
		return false, nil
	}

	// Compare the file contents with the expected string
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("error reading fee recipient file: %w", err)
	}
	existingString := string(bytes)
	if existingString != expectedString {
		// Rewrite the file with the expected distributor address
		bytes := []byte(expectedString)
		err = ioutil.WriteFile(path, bytes, FileMode)
		if err != nil {
			return false, fmt.Errorf("error writing fee recipient file: %w", err)
		}

		// If it wrote properly, indicate a success but that the file needed to be updated
		return false, nil
	}

	// The file existed and had the expected address, all set.
	return true, nil

}
