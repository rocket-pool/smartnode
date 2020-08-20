package node

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "os/exec"
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
    response := cliutils.Prompt("Would you like to detect your timezone automatically? [y/n]", "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
    fmt.Println("")
    if strings.ToLower(response[:1]) == "y" {

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
        response := cliutils.Prompt(fmt.Sprintf("The detected timezone is '%s', would you like to register using this timezone? [y/n]", timezone), "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
        fmt.Println("")
        if strings.ToLower(response[:1]) == "n" {
            timezone = ""
        }
    }

    // Prompt for time zone
    for timezone == "" {
        timezone = cliutils.Prompt("Please enter a timezone to register with in the format 'Country/City':", "^\\w{2,}\\/\\w{2,}$", "Please enter a timezone in the format 'Country/City'")
        fmt.Println("")
        response := cliutils.Prompt(fmt.Sprintf("You have chosen to register with the timezone '%s', is this correct? [y/n]", timezone), "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
        fmt.Println("")
        if strings.ToLower(response[:1]) == "n" {
            timezone = ""
        }
    }

    // Return
    return timezone

}

