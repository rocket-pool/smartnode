package teku

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
	DirMode  fs.FileMode = 0700
)

type FeeRecipientManager struct {
	keystore keystore.Keystore
}

type FeeRecipientFileContents struct {
	DefaultConfig  ProposerFeeRecipient            `json:"default_config"`
	ProposerConfig map[string]ProposerFeeRecipient `json:"proposer_config"`
}

type ProposerFeeRecipient struct {
	FeeRecipient          string                `json:"fee_recipient"`
	ValidatorRegistration ValidatorRegistration `json:"validator_registration"`
}

type ValidatorRegistration struct {
	Enabled bool `json:"enabled"`
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
	path := filepath.Join(fm.keystore.GetKeystoreDir(), config.TekuFeeRecipientFilename)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, false, nil
	} else if err != nil {
		return false, false, err
	}

	// Create the expected structure
	distributorAddress := distributor.Hex()
	expectedStruct := FeeRecipientFileContents{
		DefaultConfig: ProposerFeeRecipient{
			FeeRecipient: distributorAddress,
		},
		ProposerConfig: map[string]ProposerFeeRecipient{},
	}

	// Compare the file contents with the expected string
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return false, false, fmt.Errorf("error reading fee recipient file: %w", err)
	}
	existingStruct := &FeeRecipientFileContents{}
	err = json.Unmarshal(bytes, existingStruct)
	if err != nil {
		return false, false, fmt.Errorf("error deserializing fee recipient JSON: %w", err)
	}
	if existingStruct.DefaultConfig.FeeRecipient != expectedStruct.DefaultConfig.FeeRecipient || len(existingStruct.ProposerConfig) > 0 {
		return true, false, nil
	}
	if !existingStruct.DefaultConfig.ValidatorRegistration.Enabled {
		return true, false, nil
	}

	// The file existed and had the expected address, all set.
	return true, true, nil

}

// Writes the given address to the fee recipient file. The VC should be restarted to pick up the new file.
func (fm *FeeRecipientManager) UpdateFeeRecipientFile(distributor common.Address) error {

	// Create the expected structure
	distributorAddress := distributor.Hex()
	expectedStruct := FeeRecipientFileContents{
		DefaultConfig: ProposerFeeRecipient{
			FeeRecipient: distributorAddress,
			ValidatorRegistration: ValidatorRegistration{
				Enabled: true,
			},
		},
		ProposerConfig: map[string]ProposerFeeRecipient{},
	}
	bytes, err := json.Marshal(expectedStruct)
	if err != nil {
		return fmt.Errorf("error serializing file contents to JSON: %w", err)
	}

	// Create keystore dir
	if err := os.MkdirAll(fm.keystore.GetKeystoreDir(), DirMode); err != nil {
		return fmt.Errorf("Could not create fee recipient folder [%s]: %w", fm.keystore.GetKeystoreDir(), err)
	}

	// Write the file
	path := filepath.Join(fm.keystore.GetKeystoreDir(), config.TekuFeeRecipientFilename)
	err = ioutil.WriteFile(path, bytes, FileMode)
	if err != nil {
		return fmt.Errorf("error writing fee recipient file: %w", err)
	}
	return nil

}
