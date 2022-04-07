package prysm

import (
	"encoding/json"
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

type FeeRecipientFileContents struct {
	DefaultConfig  ProposerFeeRecipient            `json:"default_config"`
	ProposerConfig map[string]ProposerFeeRecipient `json:"proposer_config"`
}

type ProposerFeeRecipient struct {
	FeeRecipient string `json:"fee_recipient"`
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

	// Create the expected structure
	distributorAddress := distributor.Hex()
	expectedStruct := FeeRecipientFileContents{
		DefaultConfig: ProposerFeeRecipient{
			FeeRecipient: distributorAddress,
		},
		ProposerConfig: map[string]ProposerFeeRecipient{},
	}

	// Check if the file exists, and write it if it doesn't
	path := filepath.Join(fm.keystore.GetKeystoreDir(), config.PrysmFeeRecipientFilename)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		bytes, err := json.Marshal(expectedStruct)
		if err != nil {
			return false, fmt.Errorf("error serializing file contents to JSON: %w", err)
		}
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
	existingStruct := &FeeRecipientFileContents{}
	err = json.Unmarshal(bytes, existingStruct)
	if err != nil {
		return false, fmt.Errorf("error deserializing fee recipient JSON: %w", err)
	}
	if existingStruct.DefaultConfig.FeeRecipient != expectedStruct.DefaultConfig.FeeRecipient || len(existingStruct.ProposerConfig) > 0 {
		// Rewrite the file with the expected distributor address
		bytes, err := json.Marshal(expectedStruct)
		if err != nil {
			return false, fmt.Errorf("error serializing file contents to JSON: %w", err)
		}
		err = ioutil.WriteFile(path, bytes, FileMode)
		if err != nil {
			return false, fmt.Errorf("error writing fee recipient file: %w", err)
		}

		// If it wrote properly, indicate a success but that the file needed to be updated
		return false, nil
	}

	// The file existed and had the expected address, all set.
	// TODO: WAIT FOR PRYSM TO ADD SUPPORT FOR THIS, SEE
	// https://github.com/prysmaticlabs/prysm/pull/10312
	return true, fmt.Errorf("Prysm currently does not provide support for per-validator fee recipient specification, so it cannot be used to test the Merge. We will re-enable it when it has support for this feature.")

}
