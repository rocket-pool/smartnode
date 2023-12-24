package node

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/goccy/go-json"

	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
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
	if cliutils.Confirm("Would you like to detect your timezone automatically?") {
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
		if !cliutils.Confirm(fmt.Sprintf("The detected timezone is '%s', would you like to register using this timezone?", timezone)) {
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
				for _, invalidCountry := range invalidCountries {
					if invalidCountry == filename {
						isValid = false
						break
					}
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
			timezone = cliutils.Prompt("Please enter a timezone to register with in the format 'Country/City' (use Etc/UTC if you prefer not to answer):", "^([a-zA-Z_]{2,}\\/)+[a-zA-Z_]{2,}$", "Please enter a timezone in the format 'Country/City' (use Etc/UTC if you prefer not to answer)")
			if !cliutils.Confirm(fmt.Sprintf("You have chosen to register with the timezone '%s', is this correct?", timezone)) {
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
		country = cliutils.Prompt("Please enter a country / continent from the list above:", "^.+$", "Please enter a country / continent from the list above:")

		exists := false
		for _, candidate := range countryNames {
			if candidate == country {
				exists = true
				break
			}
		}

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
		region = cliutils.Prompt("Please enter a region from the list above:", "^.+$", "Please enter a region from the list above:")

		exists := false
		for _, candidate := range regionNames {
			if candidate == region {
				exists = true
				break
			}
		}

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
