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
	"time"

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
	var filter string

	// Prompt for auto-detect
	if cliutils.Confirm("Would you like to detect your timezone automatically?") {
		// Detect using the IPInfo API
		resp, err := http.Get(IPInfoURL)
		if err == nil {
			defer func() {
				_ = resp.Body.Close()
			}()
			body, err := ioutil.ReadAll(resp.Body)
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

	// Prompt for continent
	for filter == "" {
		timezone = ""
		filter = cliutils.Prompt("Enter part of the timezone (continent, country or city) to see list of options:", "^.+$", filter)

		// Gets timezones matching the provided continent removing the text until the first '/'
		cmd := fmt.Sprintf("timedatectl list-timezones --no-pager | grep '%s' ", filter)
		timezoneList, err := exec.Command("bash", "-c", cmd).Output()
		if err != nil {
			fmt.Println("Error running timedatectl:", err)
		}

		// Split the timezones
		timezones := strings.Split(string(timezoneList), "\n")

		// Print the list separated by ", "
		fmt.Println(strings.Join(timezones, ", "))

		// Prompt for the timezone
		for timezone == "" {
			timezone = cliutils.Prompt("\nPlease enter a timezone from the list in the format (Country/City) to register with (use Etc/UTC if you prefer not to answer):", "^([a-zA-Z_]{2,}\\/)+[a-zA-Z_]{2,}$", "Please enter a timezone from the list in the format (Country/City) to register with (use Etc/UTC if you prefer not to answer):")
			if !cliutils.Confirm(fmt.Sprintf("You have chosen to register with the timezone '%s', is this correct?", timezone)) {
				filter = ""
			}
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
