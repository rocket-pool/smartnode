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
            defer resp.Body.Close()
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
        timezone = cliutils.Prompt("Please enter a timezone to register with in the format 'Country/City':", "^\\w{2,}\\/\\w{2,}$", "Please enter a timezone in the format 'Country/City'")
        if !cliutils.Confirm(fmt.Sprintf("You have chosen to register with the timezone '%s', is this correct?", timezone)) {
            timezone = ""
        }
    }

    // Return
    return timezone

}


// Prompt user for a minimum node fee
func promptMinNodeFee(currentNodeFee, suggestedMinNodeFee float64) float64 {

    // Prompt for suggested min node fee
    fmt.Printf("The current network node commission rate is %f%%.\n", currentNodeFee * 100)
    fmt.Printf("The suggested minimum node commission rate for your deposit is %f%%.\n", suggestedMinNodeFee * 100)
    if cliutils.Confirm("Do you want to use the suggested minimum?") {
        return suggestedMinNodeFee
    }

    // Prompt for custom min node fee
    for {
        minNodeFeePercentStr := cliutils.Prompt("Please enter a minimum node commission rate %% for your deposit:", "^\\d+(\\.\\d+)?$", "Invalid commission rate")
        minNodeFeePercent, _ := strconv.ParseFloat(minNodeFeePercentStr, 64)
        minNodeFee := minNodeFeePercent / 100
        if minNodeFee < 0 || minNodeFee > 1 {
            fmt.Println("Invalid commission rate")
            fmt.Println("")
            continue
        }
        if cliutils.Confirm(fmt.Sprintf("You have chosen a minimum node commission rate of %f%%, is this correct?", minNodeFee * 100)) {
            return minNodeFee
        }
    }

}

