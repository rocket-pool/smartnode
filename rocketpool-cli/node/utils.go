package node

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/goccy/go-json"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"

	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	hexutils "github.com/rocket-pool/smartnode/shared/utils/hex"
)

// IPInfo API
const IPInfoURL = "https://ipinfo.io/json/"

// IPInfo response
type ipInfoResponse struct {
	Timezone string `json:"timezone"`
}

// Prompt user for a time zone string
func promptTimezone() string {

	// Time zone value
	var timezone string

	// Prompt for auto-detect
	if prompt.Confirm("Would you like to detect your timezone automatically?") {
		// Detect using the IPInfo API
		resp, err := http.Get(IPInfoURL)
		if err == nil {
			defer func() {
				_ = resp.Body.Close()
			}()
			body, err := io.ReadAll(resp.Body)
			if err == nil {
				message := new(ipInfoResponse)
				err := json.Unmarshal(body, message)
				if err == nil {
					timezone = message.Timezone
				} else {
					fmt.Printf("WARNING: couldn't query %s for your timezone based on your IP address (%s).\nChecking your system's timezone...\n", IPInfoURL, err.Error())
				}
			} else {
				fmt.Printf("WARNING: couldn't query %s for your timezone based on your IP address (%s).\nChecking your system's timezone...\n", IPInfoURL, err.Error())
			}
		} else {
			fmt.Printf("WARNING: couldn't query %s for your timezone based on your IP address (%s).\nChecking your system's timezone...\n", IPInfoURL, err.Error())
		}

		// Fall back to system time zone
		if timezone == "" {
			_, err := os.Stat("/etc/timezone")
			if os.IsNotExist(err) {
				// Try /etc/localtime, which Redhat-based systems use instead
				_, err = os.Stat("/etc/localtime")
				if err != nil {
					fmt.Printf("WARNING: couldn't get system timezone info (%s), you'll have to set it manually.\n", err.Error())
				} else {
					path, err := filepath.EvalSymlinks("/etc/localtime")
					if err != nil {
						fmt.Printf("WARNING: couldn't get system timezone info (%s), you'll have to set it manually.\n", err.Error())
					} else {
						path = strings.TrimPrefix(path, "/usr/share/zoneinfo/")
						path = strings.TrimPrefix(path, "posix/")
						path = strings.TrimSpace(path)
						// Verify it
						_, err = time.LoadLocation(path)
						if err != nil {
							fmt.Printf("WARNING: couldn't get system timezone info (%s), you'll have to set it manually.\n", err.Error())
						} else {
							timezone = path
						}
					}
				}
			} else if err != nil {
				fmt.Printf("WARNING: couldn't get system timezone info (%s), you'll have to set it manually.\n", err.Error())
			} else {
				// Debian systems
				bytes, err := os.ReadFile("/etc/timezone")
				if err != nil {
					fmt.Printf("WARNING: couldn't get system timezone info (%s), you'll have to set it manually.\n", err.Error())
				} else {
					timezone = strings.TrimSpace(string(bytes))
					// Verify it
					_, err = time.LoadLocation(timezone)
					if err != nil {
						fmt.Printf("WARNING: couldn't get system timezone info (%s), you'll have to set it manually.\n", err.Error())
						timezone = ""
					}
				}
			}
		}

	}

	// Confirm detected time zone
	if timezone != "" {
		if !prompt.Confirm(fmt.Sprintf("The detected timezone is '%s', would you like to register using this timezone?", timezone)) {
			timezone = ""
		} else {
			return timezone
		}
	}

	// Get the list of valid countries
	platformZoneSources := []string{
		"/usr/share/zoneinfo/",
		"/usr/share/lib/zoneinfo/",
		"/usr/lib/locale/TZ/",
	}
	invalidCountries := []string{
		"SystemV",
	}

	countryNames := []string{}
	for _, source := range platformZoneSources {
		files, err := os.ReadDir(source)
		if err != nil {
			continue
		}

		for _, file := range files {
			fileInfo, err := file.Info()
			if err != nil {
				continue
			}
			filename := fileInfo.Name()
			isSymlink := fileInfo.Mode()&os.ModeSymlink == os.ModeSymlink // Don't allow symlinks, which are just TZ aliases
			isDir := fileInfo.IsDir()                                     // Must be a directory
			isUpper := unicode.IsUpper(rune(filename[0]))                 // Must start with an upper case letter
			if !isSymlink && isDir && isUpper {
				isValid := true
				if slices.Contains(invalidCountries, filename) {
					isValid = false
				}
				if isValid {
					countryNames = append(countryNames, filename)
				}
			}
		}
	}

	fmt.Println("You will now be prompted to enter a timezone.")
	fmt.Println("For a complete list of valid entries, please use one of the \"TZ database name\" entries listed here:")
	fmt.Println("https://en.wikipedia.org/wiki/List_of_tz_database_time_zones")
	fmt.Println()

	// Handle situations where we couldn't parse any timezone info from the OS
	if len(countryNames) == 0 {
		for timezone == "" {
			timezone = prompt.Prompt("Please enter a timezone to register with in the format 'Country/City' (use Etc/UTC if you prefer not to answer):", "^([a-zA-Z_]{2,}\\/)+[a-zA-Z_]{2,}$", "Please enter a timezone in the format 'Country/City' (use Etc/UTC if you prefer not to answer)")
			if !prompt.Confirm(fmt.Sprintf("You have chosen to register with the timezone '%s', is this correct?", timezone)) {
				timezone = ""
			}
		}

		// Return
		return timezone
	}

	// Print countries
	sort.Strings(countryNames)
	fmt.Println("List of valid countries / continents:")
	for _, countryName := range countryNames {
		fmt.Println(countryName)
	}
	fmt.Println()

	// Prompt for country
	country := ""
	for {
		time.Now().Zone()
		timezone = ""
		country = prompt.Prompt("Please enter a country / continent from the list above:", "^.+$", "Please enter a country / continent from the list above:")

		exists := slices.Contains(countryNames, country)

		if !exists {
			fmt.Printf("%s is not a valid country or continent. Please see the list above for valid countries and continents.\n\n", country)
		} else {
			break
		}
	}

	// Get the list of regions for the selected country
	regionNames := []string{}
	for _, source := range platformZoneSources {
		files, err := os.ReadDir(filepath.Join(source, country))
		if err != nil {
			continue
		}

		for _, file := range files {
			fileInfo, err := file.Info()
			if err != nil {
				continue
			}
			if fileInfo.IsDir() {
				subfiles, err := os.ReadDir(filepath.Join(source, country, fileInfo.Name()))
				if err != nil {
					continue
				}
				for _, subfile := range subfiles {
					subfileInfo, err := subfile.Info()
					if err != nil {
						continue
					}
					regionNames = append(regionNames, fmt.Sprintf("%s/%s", fileInfo.Name(), subfileInfo.Name()))
				}
			} else {
				regionNames = append(regionNames, fileInfo.Name())
			}
		}
	}

	// Print regions
	sort.Strings(regionNames)
	fmt.Println("List of valid regions:")
	for _, regionName := range regionNames {
		fmt.Println(regionName)
	}
	fmt.Println()

	// Prompt for region
	region := ""
	for {
		time.Now().Zone()
		timezone = ""
		region = prompt.Prompt("Please enter a region from the list above:", "^.+$", "Please enter a region from the list above:")

		exists := slices.Contains(regionNames, region)

		if !exists {
			fmt.Printf("%s is not a valid country or continent. Please see the list above for valid countries and continents.\n\n", region)
		} else {
			break
		}
	}

	// Return
	timezone = fmt.Sprintf("%s/%s", country, region)
	fmt.Printf("Using timezone %s.\n", timezone)
	return timezone
}

// Prompt user for a minimum node fee
func promptMinNodeFee(networkCurrentNodeFee, networkMinNodeFee float64) float64 {

	// Get suggested min node fee
	suggestedMinNodeFee := networkCurrentNodeFee - defaultMaxNodeFeeSlippage
	if suggestedMinNodeFee < networkMinNodeFee {
		suggestedMinNodeFee = networkMinNodeFee
	}

	// Prompt for suggested max slippage
	fmt.Printf("The current network node commission rate that your minipool should receive is %f%%.\n", networkCurrentNodeFee*100)
	fmt.Printf("The suggested maximum commission rate slippage for your deposit transaction is %f%%.\n", defaultMaxNodeFeeSlippage*100)
	fmt.Printf("This will result in your minipool receiving a minimum possible commission rate of %f%%.\n", suggestedMinNodeFee*100)
	if prompt.Confirm("Do you want to use the suggested maximum commission rate slippage?") {
		return suggestedMinNodeFee
	}

	// Prompt for custom max slippage
	for {

		// Get max slippage
		maxNodeFeeSlippagePercStr := prompt.Prompt("Please enter a maximum commission rate slippage % for your deposit:", "^\\d+(\\.\\d+)?$", "Invalid maximum commission rate slippage")
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
		if prompt.Confirm(fmt.Sprintf("You have chosen a maximum commission rate slippage of %f%%, resulting in a minimum possible commission rate of %f%%. Is this correct?", maxNodeFeeSlippage*100, minNodeFee*100)) {
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
	files, err := os.ReadDir(customKeyDir)
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
		bytes, err := os.ReadFile(filepath.Join(customKeyDir, file.Name()))
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
			password := prompt.PromptPassword(
				fmt.Sprintf("Please enter the password that the keystore for %s was encrypted with:", pubkey.Hex()), "^.*$", "",
			)

			formattedPubkey := strings.ToUpper(hexutils.RemovePrefix(pubkey.Hex()))
			pubkeyPasswords[formattedPubkey] = password

			fmt.Println()
			break
		}
	}

	if len(pubkeyPasswords) == 0 {
		return "", fmt.Errorf("couldn't find the keystore for validator %s in the custom-keys directory; if you want to import this key into the Smartnode stack, you will need to put its keystore file into custom-keys first", pubkey.String())
	}

	// Store it in the file
	fileBytes, err := yaml.Marshal(pubkeyPasswords)
	if err != nil {
		return "", fmt.Errorf("error serializing keystore passwords file: %w", err)
	}
	passwordFile := filepath.Join(datapath, "custom-key-passwords")
	err = os.WriteFile(passwordFile, fileBytes, 0600)
	if err != nil {
		return "", fmt.Errorf("error writing keystore passwords file: %w", err)
	}

	return passwordFile, nil

}

// Display a warning if hotfix is live and voting is uninitialized
func warnIfVotingUninitialized(rp *rocketpool.Client, c *cli.Context, warningMessage string) error {
	// Check for Houston 1.3.1 Hotfix
	hotfix, err := rp.IsHoustonHotfixDeployed()
	if err != nil {
		return fmt.Errorf("error checking if Houston Hotfix has been deployed: %w", err)
	}

	if hotfix.IsHoustonHotfixDeployed {
		// Check if voting power is initialized
		isVotingInitializedResponse, err := rp.IsVotingInitialized()
		if err != nil {
			return fmt.Errorf("error checking if voting is initialized: %w", err)
		}
		if !isVotingInitializedResponse.VotingInitialized {
			fmt.Println("Your voting power hasn't been initialized yet. Please visit https://docs.rocketpool.net/guides/houston/participate#initializing-voting to learn more.")
			// Post a warning about initializing voting
			if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("%s%s%s\nWould you like to continue?", colorYellow, warningMessage, colorReset))) {
				fmt.Println("Cancelled.")
				return fmt.Errorf("operation cancelled by user")
			}
		}
	}

	return nil
}
