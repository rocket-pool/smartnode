package rocketpool

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"golang.org/x/sync/errgroup"
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
	keymanagerPort := cfg.KeymanagerApiPort()
	var keymanagerHost string
	if cfg.IsNativeMode {
		keymanagerHost = "127.0.0.1"
	} else {
		// Create the hostname string (example: rocketpool_validator)
		projectName := cfg.Smartnode.ProjectName.Value.(string)
		keymanagerHost = fmt.Sprintf("%s_validator", projectName)
	}
	keymanagerURL := fmt.Sprintf("http://%s:%d", keymanagerHost, keymanagerPort)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	var wg errgroup.Group
	// Iterate through megapool pubkeys and update the fee recipient for each
	for _, pubkey := range pubkeys {
		// start a goroutine for each pubkey
		wg.Go(func() error {

			pubkeyHex := fmt.Sprintf("0x%s", pubkey.Hex())

			endpoint := fmt.Sprintf("%s/eth/v1/validator/%s/feerecipient", keymanagerURL, pubkeyHex)

			getRequest, err := http.NewRequest("GET", endpoint, nil)
			if err != nil {
				return fmt.Errorf("error creating request for pubkey %s: %w", pubkeyHex, err)
			}

			getRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", KeymanagerToken))

			resp, err := client.Do(getRequest)
			if err != nil {
				return fmt.Errorf("error making GET request for pubkey %s: %w", pubkeyHex, err)
			}
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				return fmt.Errorf("error reading GET response body for pubkey %s: %w", pubkeyHex, err)

			}

			type FeeRecipientResp struct {
				Data struct {
					Ethaddress string `json:"ethaddress"`
				} `json:"data"`
			}

			var frResp FeeRecipientResp
			err = json.Unmarshal(body, &frResp)
			if err != nil {
				return fmt.Errorf("error to unmarshall fee recipient body %w", err)
			}

			if !strings.EqualFold(frResp.Data.Ethaddress, megapoolAddress.Hex()) {
				fmt.Printf("updating fee recipient for key: %s old recipient %s\n", pubkey.Hex(), frResp.Data.Ethaddress)
				postBody := map[string]string{
					"ethaddress": megapoolAddress.Hex(),
				}
				jsonBody, err := json.Marshal(postBody)
				if err != nil {
					return fmt.Errorf("error marshaling request body for pubkey %s: %w", pubkeyHex, err)
				}

				postRequest, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonBody))
				if err != nil {
					return fmt.Errorf("error creating request for pubkey %s: %w", pubkeyHex, err)
				}

				postRequest.Header.Set("Content-Type", "application/json")
				postRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", KeymanagerToken))

				resp, err := client.Do(postRequest)
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
		})

	}
	if err := wg.Wait(); err != nil {
		return err
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
