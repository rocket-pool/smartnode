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


// Test node register command
func TestNodeRegister(t *testing.T) {

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
    initArgs := test.GetAppArgs(dataPath, initInput.Name(), "")
    registerArgs := test.GetAppArgs(dataPath, registerInput.Name(), output.Name())
    appOptions := test.GetAppOptions(dataPath)

    // Attempt to register uninitialised node
    if err := app.Run(append(registerArgs, "node", "register")); err == nil { t.Error("Should return error for uninitialised node") }

    // Initialise node
    if err := app.Run(append(initArgs, "node", "init")); err != nil { t.Error(err) }

    // Register initialised node with no balance
    if err := app.Run(append(registerArgs, "node", "register")); err != nil { t.Error(err) }

    // Seed node account
    if err := test.AppSeedAccount(appOptions, eth.EthToWei(10)); err != nil { t.Fatal(err) }

    // Register initialised node with balance
    if err := app.Run(append(registerArgs, "node", "register")); err != nil { t.Error(err) }

    // Register already registered node
    if err := app.Run(append(registerArgs, "node", "register")); err != nil { t.Error(err) }

    // Read & check output
    output, err = os.Open(output.Name())
    if err != nil { t.Fatal(err) }
    line := 0
    for scanner := bufio.NewScanner(output); scanner.Scan(); {
        if regexp.MustCompile("(?i)^Your system timezone is").MatchString(scanner.Text()) || regexp.MustCompile("(?i)^Please answer").MatchString(scanner.Text()) { continue }
        line++
        switch line {
            case 1: if !regexp.MustCompile("(?i)^Node account 0x[0-9a-fA-F]{40} requires a minimum balance of \\d\\.\\d\\d ETH to operate in Rocket Pool$").MatchString(scanner.Text()) { t.Error("Minimum balance message incorrect") }
            case 2: if !regexp.MustCompile("(?i)^Registering node...$").MatchString(scanner.Text()) { t.Error("Registering node message incorrect") }
            case 3: if !regexp.MustCompile("(?i)^Node registered successfully with Rocket Pool - new node deposit contract created at 0x[0-9a-fA-F]{40}$").MatchString(scanner.Text()) { t.Error("Node registered message incorrect") }
            case 4: if !regexp.MustCompile("(?i)^Node is already registered with Rocket Pool - current deposit contract is at 0x[0-9a-fA-F]{40}$").MatchString(scanner.Text()) { t.Error("Node already registered message incorrect") }
        }
    }
    if line != 4 { t.Error("Incorrect output line count") }

}

