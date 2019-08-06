package node

import (
    "bufio"
    "io/ioutil"
    "testing"
    "os"
    "regexp"

    "github.com/rocket-pool/smartnode/shared/utils/eth"

    test "github.com/rocket-pool/smartnode/tests/utils"
)


// Test node timezone command
func TestNodeTimezone(t *testing.T) {

    // Create test app
    app := test.NewApp()

    // Create temporary input files
    initInput, err := test.NewInputFile("foobarbaz" + "\n")
    if err != nil { t.Fatal(err) }
    initInput.Close()
    timezoneInput, err := test.NewInputFile(
        "Australia/Brisbane" + "\n" +
        "YES" + "\n")
    if err != nil { t.Fatal(err) }
    timezoneInput.Close()

    // Create temporary output file
    output, err := ioutil.TempFile("", "")
    if err != nil { t.Fatal(err) }
    output.Close()

    // Create temporary data path
    dataPath, err := ioutil.TempDir("", "")
    if err != nil { t.Fatal(err) }

    // Get app args & options
    initArgs := test.GetAppArgs(dataPath, initInput.Name(), "")
    registerArgs := test.GetAppArgs(dataPath, timezoneInput.Name(), "")
    timezoneArgs := test.GetAppArgs(dataPath, timezoneInput.Name(), output.Name())
    appOptions := test.GetAppOptions(dataPath)

    // Attempt to set timezone for uninitialised node
    if err := app.Run(append(timezoneArgs, "node", "timezone")); err == nil { t.Error("Should return error for uninitialised node") }

    // Initialise node
    if err := app.Run(append(initArgs, "node", "init")); err != nil { t.Error(err) }

    // Attempt to set timezone for unregistered node
    if err := app.Run(append(timezoneArgs, "node", "timezone")); err == nil { t.Error("Should return error for unregistered node") }

    // Seed node account & register node
    if err := test.AppSeedAccount(appOptions, eth.EthToWei(10)); err != nil { t.Fatal(err) }
    if err := app.Run(append(registerArgs, "node", "register")); err != nil { t.Error(err) }

    // Set timezone for registered node
    if err := app.Run(append(timezoneArgs, "node", "timezone")); err != nil { t.Error(err) }

    // Read & check output
    output, err = os.Open(output.Name())
    if err != nil { t.Fatal(err) }
    line := 0
    for scanner := bufio.NewScanner(output); scanner.Scan(); {
        if regexp.MustCompile("(?i)^Your system timezone is").MatchString(scanner.Text()) || regexp.MustCompile("(?i)^Please answer").MatchString(scanner.Text()) { continue }
        line++
        switch line {
            case 1: if !regexp.MustCompile("(?i)^Setting node timezone...$").MatchString(scanner.Text()) { t.Error("Setting node timezone message incorrect") }
            case 2: if !regexp.MustCompile("(?i)^Node timezone successfully updated to: \\w+/\\w+$").MatchString(scanner.Text()) { t.Error("Node timezone updated message incorrect") }
        }
    }
    if line != 2 { t.Error("Incorrect output line count") }

}

