package node

import (
    "fmt"
    "os/exec"
    "regexp"
    "strings"

    cliutils "github.com/rocket-pool/smartnode-cli/rocketpool/utils/cli"
)


// Prompt user for a time zone string
func promptTimezone() string {

    // Time zone value
    var timezone string

    // Get system time zone
    timeOutput, _ := exec.Command("timedatectl").Output()
    if len(timeOutput) > 0 {
        tzMatches := regexp.MustCompile("(?i)zone:\\s*(\\w{2,}\\/\\w{2,})").FindStringSubmatch(string(timeOutput[:]))
        if len(tzMatches) > 1 {
            timezone = tzMatches[1]
        }
    }

    // Confirm system time zone
    if timezone != "" {
        response := cliutils.Prompt(fmt.Sprintf("Your system timezone is '%s', would you like to register using this timezone? [y/n]", timezone), "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
        if strings.ToLower(response[:1]) == "n" {
            timezone = ""
        }
    }

    // Prompt for time zone
    for timezone == "" {
        timezone = cliutils.Prompt("Please enter a timezone to register with in the format 'Country/City':", "^\\w{2,}\\/\\w{2,}$", "Please enter a timezone in the format 'Country/City'")
        response := cliutils.Prompt(fmt.Sprintf("You have chosen to register with the timezone '%s', is this correct? [y/n]", timezone), "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
        if strings.ToLower(response[:1]) == "n" {
            timezone = ""
        }
    }

    // Return
    return timezone

}

