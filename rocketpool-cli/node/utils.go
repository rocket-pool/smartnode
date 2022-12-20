package node

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	hexutils "github.com/rocket-pool/smartnode/shared/utils/hex"
	"gopkg.in/yaml.v2"
)

// FreeGeoIP config
const FreeGeoIPURL = "https://freegeoip.app/json/"

// FreeGeoIP response
type freeGeoIPResponse struct {
	Timezone string `json:"time_zone"`
}

// Prompt user for a time zone string
func promptTimezone() string {

	// Time zone value
	var timezone string

	// Prompt for auto-detect
	if cliutils.Confirm("Would you like to detect your timezone automatically?") {
		// Detect using FreeGeoIP
		if resp, err := http.Get(FreeGeoIPURL); err == nil {
			defer func() {
				_ = resp.Body.Close()
			}()
			if body, err := ioutil.ReadAll(resp.Body); err == nil {
				message := new(freeGeoIPResponse)
				if err := json.Unmarshal(body, message); err == nil {
					timezone = message.Timezone
				}
			}
		}

		// Fall back to system time zone
		if timezone == "" {
			if tzOutput, _ := exec.Command("cat", "/etc/timezone").Output(); len(tzOutput) > 0 {
				timezone = strings.TrimSpace(string(tzOutput))
			}
		}

	}

	// Confirm detected time zone
	if timezone != "" {
		if !cliutils.Confirm(fmt.Sprintf("The detected timezone is '%s', would you like to register using this timezone?", timezone)) {
			timezone = ""
		}
	}

	// Prompt for time zone
	for timezone == "" {
		timezone = cliutils.Prompt("Please enter a timezone to register with in the format 'Country/City' (use Etc/UTC if you prefer not to answer):", "^([a-zA-Z_]{2,}\\/)+[a-zA-Z_]{2,}$", "Please enter a timezone in the format 'Country/City' (use Etc/UTC if you prefer not to answer)")
		if !cliutils.Confirm(fmt.Sprintf("You have chosen to register with the timezone '%s', is this correct?", timezone)) {
			timezone = ""
		}
	}

	// Return
	return timezone

}

// Prompt user for a minimum node fee
func promptMinNodeFee(networkCurrentNodeFee, networkMinNodeFee float64) float64 {

	// Get suggested min node fee
	suggestedMinNodeFee := networkCurrentNodeFee - DefaultMaxNodeFeeSlippage
	if suggestedMinNodeFee < networkMinNodeFee {
		suggestedMinNodeFee = networkMinNodeFee
	}

	// Prompt for suggested max slippage
	fmt.Printf("The current network node commission rate that your minipool should receive is %f%%.\n", networkCurrentNodeFee*100)
	fmt.Printf("The suggested maximum commission rate slippage for your deposit transaction is %f%%.\n", DefaultMaxNodeFeeSlippage*100)
	fmt.Printf("This will result in your minipool receiving a minimum possible commission rate of %f%%.\n", suggestedMinNodeFee*100)
	if cliutils.Confirm("Do you want to use the suggested maximum commission rate slippage?") {
		return suggestedMinNodeFee
	}

	// Prompt for custom max slippage
	for {

		// Get max slippage
		maxNodeFeeSlippagePercStr := cliutils.Prompt("Please enter a maximum commission rate slippage % for your deposit:", "^\\d+(\\.\\d+)?$", "Invalid maximum commission rate slippage")
		maxNodeFeeSlippagePerc, _ := strconv.ParseFloat(maxNodeFeeSlippagePercStr, 64)
		maxNodeFeeSlippage := maxNodeFeeSlippagePerc / 100
		if maxNodeFeeSlippage < 0 || maxNodeFeeSlippage > 1 {
			fmt.Println("Invalid maximum commission rate slippage")
			fmt.Println("")
			continue
		}

		// Calculate min node fee
		minNodeFee := networkCurrentNodeFee - maxNodeFeeSlippage
		if minNodeFee < networkMinNodeFee {
			minNodeFee = networkMinNodeFee
		}

		// Confirm max slippage
		if cliutils.Confirm(fmt.Sprintf("You have chosen a maximum commission rate slippage of %f%%, resulting in a minimum possible commission rate of %f%%. Is this correct?", maxNodeFeeSlippage*100, minNodeFee*100)) {
			return minNodeFee
		}

	}

}

// Prompt for the password to a solo validator key as part of migration
func promptForSoloKeyPassword(rp *rocketpool.Client, cfg *config.RocketPoolConfig, pubkey types.ValidatorPubkey) (string, error) {

	// Check for the custom key directory
	datapath, err := homedir.Expand(cfg.Smartnode.DataPath.Value.(string))
	if err != nil {
		return "", fmt.Errorf("error expanding data directory: %w", err)
	}
	customKeyDir := filepath.Join(datapath, "custom-keys")
	info, err := os.Stat(customKeyDir)
	if os.IsNotExist(err) || !info.IsDir() {
		return "", nil
	}

	// Get the custom keystore files
	files, err := ioutil.ReadDir(customKeyDir)
	if err != nil {
		return "", fmt.Errorf("error enumerating custom keystores: %w", err)
	}
	if len(files) == 0 {
		return "", nil
	}

	// Get the pubkeys for the custom keystores
	pubkeyPasswords := map[string]string{}
	for _, file := range files {
		// Read the file
		bytes, err := ioutil.ReadFile(filepath.Join(customKeyDir, file.Name()))
		if err != nil {
			return "", fmt.Errorf("error reading custom keystore %s: %w", file.Name(), err)
		}

		// Deserialize it
		keystore := api.ValidatorKeystore{}
		err = json.Unmarshal(bytes, &keystore)
		if err != nil {
			return "", fmt.Errorf("error deserializing custom keystore %s: %w", file.Name(), err)
		}

		if keystore.Pubkey == pubkey {
			// Found it, prompt for the password
			password := cliutils.PromptPassword(
				fmt.Sprintf("Please enter the password that the keystore for %s was encrypted with:", pubkey.Hex()), "^.*$", "",
			)

			formattedPubkey := strings.ToUpper(hexutils.RemovePrefix(pubkey.Hex()))
			pubkeyPasswords[formattedPubkey] = password

			fmt.Println()
			break
		}
	}

	if len(pubkeyPasswords) == 0 {
		return "", fmt.Errorf("couldn't find the keystore for validator %s in the custom-keys directory; if you want to import this key into the Smartnode stack, you will need to put its keystore file into custom-keys first")
	}

	// Store it in the file
	fileBytes, err := yaml.Marshal(pubkeyPasswords)
	if err != nil {
		return "", fmt.Errorf("error serializing keystore passwords file: %w", err)
	}
	passwordFile := filepath.Join(datapath, "custom-key-passwords")
	err = ioutil.WriteFile(passwordFile, fileBytes, 0600)
	if err != nil {
		return "", fmt.Errorf("error writing keystore passwords file: %w", err)
	}

	return passwordFile, nil

}
