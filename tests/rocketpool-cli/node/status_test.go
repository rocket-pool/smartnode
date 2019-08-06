package node

import (
    "bufio"
    "io/ioutil"
    "testing"
    "os"
    "regexp"

    "github.com/rocket-pool/smartnode/shared/utils/eth"

    test "github.com/rocket-pool/smartnode/tests/utils"
    rp "github.com/rocket-pool/smartnode/tests/utils/rocketpool"
)


// Test node status command
func TestNodeStatus(t *testing.T) {

    // Create test app
    app := test.NewApp()

    // Create temporary input files
    initInput, err := test.NewInputFile("foobarbaz" + "\n")
    if err != nil { t.Fatal(err) }
    initInput.Close()
    registerInput, err := test.NewInputFile(
        "Australia/Brisbane" + "\n" +
        "YES" + "\n")
    if err != nil { t.Fatal(err) }
    registerInput.Close()

    // Create temporary output file
    output, err := ioutil.TempFile("", "")
    if err != nil { t.Fatal(err) }
    output.Close()

    // Create temporary data path
    dataPath, err := ioutil.TempDir("", "")
    if err != nil { t.Fatal(err) }

    // Get app args & options
    statusArgs := test.GetAppArgs(dataPath, "", output.Name())
    initArgs := test.GetAppArgs(dataPath, initInput.Name(), "")
    registerArgs := test.GetAppArgs(dataPath, registerInput.Name(), "")
    appOptions := test.GetAppOptions(dataPath)

    // Attempt to get status of uninitialised node
    if err := app.Run(append(statusArgs, "node", "status")); err == nil { t.Error("Should return error for uninitialised node") }

    // Initialise node
    if err := app.Run(append(initArgs, "node", "init")); err != nil { t.Error(err) }
    
    // Get status of unregistered node
    if err := app.Run(append(statusArgs, "node", "status")); err != nil { t.Error(err) }

    // Seed node account & register node
    if err := test.AppSeedAccount(appOptions, eth.EthToWei(10)); err != nil { t.Fatal(err) }
    if err := app.Run(append(registerArgs, "node", "register")); err != nil { t.Error(err) }

    // Get status of registered node
    if err := app.Run(append(statusArgs, "node", "status")); err != nil { t.Error(err) }

    // Make node trusted
    if err := rp.AppSetNodeTrusted(appOptions); err != nil { t.Fatal(err) }

    // Get status of trusted node
    if err := app.Run(append(statusArgs, "node", "status")); err != nil { t.Error(err) }

    // Read & check output
    output, err = os.Open(output.Name())
    if err != nil { t.Fatal(err) }
    line := 0
    for scanner := bufio.NewScanner(output); scanner.Scan(); {
        line++
        switch line {
            case 1: fallthrough
            case 3: fallthrough
            case 5: if !regexp.MustCompile("(?i)^Node account 0x[0-9a-fA-F]{40} has a balance of \\d\\.\\d\\d ETH, \\d\\.\\d\\d rETH and \\d\\.\\d\\d RPL$").MatchString(scanner.Text()) { t.Error("Node account message incorrect") }
            case 2: if !regexp.MustCompile("(?i)^Node is not registered with Rocket Pool$").MatchString(scanner.Text()) { t.Error("Node not registered message incorrect") }
            case 4: fallthrough
            case 6: if !regexp.MustCompile("(?i)^Node registered with Rocket Pool with contract at 0x[0-9a-fA-F]{40}, timezone '\\w+/\\w+' and a balance of \\d\\.\\d\\d ETH and \\d\\.\\d\\d RPL$").MatchString(scanner.Text()) { t.Error("Node registered message incorrect") }
            case 7: if !regexp.MustCompile("(?i)^Node is a trusted Rocket Pool node and will perform watchtower duties$").MatchString(scanner.Text()) { t.Error("Node trusted message incorrect") }
        }
    }
    if line != 7 { t.Error("Incorrect output line count") }

}

