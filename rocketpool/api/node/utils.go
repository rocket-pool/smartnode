package node

import (
    "fmt"
    "os/exec"
    "regexp"
    "strings"

    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/cli"
)


// Prompt user for a time zone string
func promptTimezone() string {

    // Time zone value
    var timezone string

    // Get system time zone
    if timeOutput, _ := exec.Command("timedatectl").Output(); len(timeOutput) > 0 {
        if tzMatches := regexp.MustCompile("(?i)zone:\\s*(\\w{2,}\\/\\w{2,})").FindStringSubmatch(string(timeOutput[:])); len(tzMatches) > 1 {
            timezone = tzMatches[1]
        }
    }

    // Confirm system time zone
    if timezone != "" {
        response := cli.Prompt(fmt.Sprintf("Your system timezone is '%s', would you like to register using this timezone? [y/n]", timezone), "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
        if strings.ToLower(response[:1]) == "n" {
            timezone = ""
        }
    }

    // Prompt for time zone
    for timezone == "" {
        timezone = cli.Prompt("Please enter a timezone to register with in the format 'Country/City':", "^\\w{2,}\\/\\w{2,}$", "Please enter a timezone in the format 'Country/City'")
        response := cli.Prompt(fmt.Sprintf("You have chosen to register with the timezone '%s', is this correct? [y/n]", timezone), "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
        if strings.ToLower(response[:1]) == "n" {
            timezone = ""
        }
    }

    // Return
    return timezone

}

