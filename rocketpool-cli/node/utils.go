package node

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// IP locator API config
const GeolocatorIPURL = "https://ipinfo.io/json"

// IP locator API response
type geolocatorIPResponse struct {
	Timezone string `json:"timezone"`
}

// Prompt user for a time zone string
func promptTimezone() string {

	// Time zone value
	var timezone string

	// Prompt for auto-detect
	if cliutils.Confirm("Would you like to detect your timezone automatically?") {
		// Detect using IP locator API
		if resp, err := http.Get(GeolocatorIPURL); err == nil {
			defer func() {
				_ = resp.Body.Close()
			}()
			if body, err := ioutil.ReadAll(resp.Body); err == nil {
				message := new(geolocatorIPResponse)
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
