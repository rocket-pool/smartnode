package node

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
    "os/exec"
    "strings"

    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


// FreeGeoIP config
const FREE_GEO_IP_URL string = "https://freegeoip.app/json/"


// FreeGeoIP response
type FreeGeoIPResponse struct {
    Timezone string `json:"time_zone"`
}


// Prompt user for a time zone string
func promptTimezone(input *os.File, output *os.File) string {

    // Time zone value
    var timezone string

    // Prompt for auto-detect
    response := cliutils.Prompt(input, output, "Would you like to detect your timezone automatically? [y/n]", "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
    if strings.ToLower(response[:1]) == "y" {

        // Detect using FreeGeoIP
        if resp, err := http.Get(FREE_GEO_IP_URL); err == nil {
            defer resp.Body.Close()
            if body, err := ioutil.ReadAll(resp.Body); err == nil {
                message := new(FreeGeoIPResponse)
                if err := json.Unmarshal(body, message); err == nil {
                    timezone = message.Timezone
                }
            }
        }

        // Fall back to system time zone
        if timezone == "" {
            if tzOutput, _ := exec.Command("cat", "/etc/timezone").Output(); len(tzOutput) > 0 {
                timezone = strings.TrimSpace(string(tzOutput[:]))
            }
        }

    }

    // Confirm system time zone
    if timezone != "" {
        response := cliutils.Prompt(input, output, fmt.Sprintf("The detected timezone is '%s', would you like to register using this timezone? [y/n]", timezone), "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
        if strings.ToLower(response[:1]) == "n" {
            timezone = ""
        }
    }

    // Prompt for time zone
    for timezone == "" {
        timezone = cliutils.Prompt(input, output, "Please enter a timezone to register with in the format 'Country/City':", "^\\w{2,}\\/\\w{2,}$", "Please enter a timezone in the format 'Country/City'")
        response := cliutils.Prompt(input, output, fmt.Sprintf("You have chosen to register with the timezone '%s', is this correct? [y/n]", timezone), "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
        if strings.ToLower(response[:1]) == "n" {
            timezone = ""
        }
    }

    // Return
    return timezone

}

