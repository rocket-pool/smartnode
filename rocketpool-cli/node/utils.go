package node

import (
    "fmt"
    "os"
    "os/exec"
    "strings"

    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


// Prompt user for a time zone string
func promptTimezone(input *os.File) string {

    // Time zone value
    var timezone string

    // Get system time zone
    if tzOutput, _ := exec.Command("cat", "/etc/timezone").Output(); len(tzOutput) > 0 {
        timezone = strings.TrimSpace(string(tzOutput[:]))
    }

    // Confirm system time zone
    if timezone != "" {
        response := cliutils.Prompt(input, fmt.Sprintf("Your system timezone is '%s', would you like to register using this timezone? [y/n]", timezone), "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
        if strings.ToLower(response[:1]) == "n" {
            timezone = ""
        }
    }

    // Prompt for time zone
    for timezone == "" {
        timezone = cliutils.Prompt(input, "Please enter a timezone to register with in the format 'Country/City':", "^\\w{2,}\\/\\w{2,}$", "Please enter a timezone in the format 'Country/City'")
        response := cliutils.Prompt(input, fmt.Sprintf("You have chosen to register with the timezone '%s', is this correct? [y/n]", timezone), "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
        if strings.ToLower(response[:1]) == "n" {
            timezone = ""
        }
    }

    // Return
    return timezone

}

