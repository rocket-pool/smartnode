package rocketpool

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/shared/services/config"
)

// Config
const (
	FileMode        fs.FileMode = 0644
	KeymanagerToken             = "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
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

// Iterates pubkeys making keymanager API calls updating the fee recipient
func UpdateFeeRecipientPerKey(pubkeys []types.ValidatorPubkey, megapoolAddress common.Address, cfg *config.RocketPoolConfig) error {
	if len(pubkeys) == 0 {
		return nil
	}

	// Get the keymanager API URL
	keymanagerPort := cfg.ConsensusCommon.KeymanagerApiPort.Value.(uint16)
	keymanagerURL := fmt.Sprintf("http://127.0.0.1:%d", keymanagerPort)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Iterate through megapool pubkeys and update the fee recipient for each
	for _, pubkey := range pubkeys {
		pubkeyHex := pubkey.Hex()

		endpoint := fmt.Sprintf("%s/eth/v1/validator/%s/feerecipient", keymanagerURL, pubkeyHex)

		requestBody := map[string]string{
			"ethaddress": megapoolAddress.Hex(),
		}
		jsonBody, err := json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("error marshaling request body for pubkey %s: %w", pubkeyHex, err)
		}

		req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonBody))
		if err != nil {
			return fmt.Errorf("error creating request for pubkey %s: %w", pubkeyHex, err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", KeymanagerToken))

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("error making request for pubkey %s: %w", pubkeyHex, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			bodyBytes, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("keymanager API error for pubkey %s: %s", pubkeyHex, string(bodyBytes))
		}
	}

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
