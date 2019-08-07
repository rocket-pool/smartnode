package node

import (
    "io/ioutil"
    "testing"

    "github.com/rocket-pool/smartnode/shared/utils/eth"

    test "github.com/rocket-pool/smartnode/tests/utils"
    testapp "github.com/rocket-pool/smartnode/tests/utils/app"
)


// Test node timezone command
func TestNodeTimezone(t *testing.T) {

    // Create test app
    app := testapp.NewApp()

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
    timezoneArgs := testapp.GetAppArgs(dataPath, timezoneInput.Name(), output.Name())
    initArgs := testapp.GetAppArgs(dataPath, initInput.Name(), "")
    registerArgs := testapp.GetAppArgs(dataPath, timezoneInput.Name(), "")
    appOptions := testapp.GetAppOptions(dataPath)

    // Attempt to set timezone for uninitialised node
    if err := app.Run(append(timezoneArgs, "node", "timezone")); err == nil { t.Error("Should return error for uninitialised node") }

    // Initialise node
    if err := app.Run(append(initArgs, "node", "init")); err != nil { t.Fatal(err) }

    // Attempt to set timezone for unregistered node
    if err := app.Run(append(timezoneArgs, "node", "timezone")); err == nil { t.Error("Should return error for unregistered node") }

    // Seed node account & register node
    if err := testapp.AppSeedNodeAccount(appOptions, eth.EthToWei(10), nil); err != nil { t.Fatal(err) }
    if err := app.Run(append(registerArgs, "node", "register")); err != nil { t.Fatal(err) }

    // Set timezone for registered node
    if err := app.Run(append(timezoneArgs, "node", "timezone")); err != nil { t.Error(err) }

    // Check output
    if messages, err := testapp.CheckOutput(output.Name(), []string{"(?i)^Your system timezone is", "(?i)^Please answer"}, map[int][]string{
        1: []string{"(?i)^Setting node timezone...$", "Setting node timezone message incorrect"},
        2: []string{"(?i)^Node timezone successfully updated to: \\w+/\\w+$", "Node timezone updated message incorrect"},
    }); err != nil {
        t.Fatal(err)
    } else {
        for _, msg := range messages { t.Error(msg) }
    }

}

