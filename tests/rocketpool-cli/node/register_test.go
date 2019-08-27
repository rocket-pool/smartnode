package node

import (
    "io/ioutil"
    "testing"

    "github.com/rocket-pool/smartnode/shared/utils/eth"

    test "github.com/rocket-pool/smartnode/tests/utils"
    testapp "github.com/rocket-pool/smartnode/tests/utils/app"
)


// Test node register command
func TestNodeRegister(t *testing.T) {

    // Create test app
    app := testapp.NewApp()

    // Create temporary input files
    initInput, err := test.NewInputFile("foobarbaz" + "\n")
    if err != nil { t.Fatal(err) }
    initInput.Close()
    registerInput, err := test.NewInputFile(
        "NO" + "\n" +
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
    initArgs := testapp.GetAppArgs(dataPath, initInput.Name(), "")
    registerArgs := testapp.GetAppArgs(dataPath, registerInput.Name(), output.Name())
    appOptions := testapp.GetAppOptions(dataPath)

    // Attempt to register uninitialised node
    if err := app.Run(append(registerArgs, "node", "register")); err == nil { t.Error("Should return error for uninitialised node") }

    // Initialise node
    if err := app.Run(append(initArgs, "node", "init")); err != nil { t.Fatal(err) }

    // Register initialised node with no balance
    if err := app.Run(append(registerArgs, "node", "register")); err != nil { t.Error(err) }

    // Seed node account
    if err := testapp.AppSeedNodeAccount(appOptions, eth.EthToWei(10), nil); err != nil { t.Fatal(err) }

    // Register initialised node with balance
    if err := app.Run(append(registerArgs, "node", "register")); err != nil { t.Error(err) }

    // Register already registered node
    if err := app.Run(append(registerArgs, "node", "register")); err != nil { t.Error(err) }

    // Check output
    if messages, err := testapp.CheckOutput(output.Name(), []string{"(?i)^Would you like to detect your timezone", "(?i)^Please enter a timezone", "(?i)^You have chosen to register with the timezone", "(?i)^Registering node"}, map[int][]string{
        1: []string{"(?i)^Node account 0x[0-9a-fA-F]{40} requires a minimum balance of \\d+\\.\\d+ ETH to operate in Rocket Pool$", "Minimum balance message incorrect"},
        2: []string{"(?i)^Node registered successfully with Rocket Pool - new node deposit contract created at 0x[0-9a-fA-F]{40}$", "Node registered message incorrect"},
        3: []string{"(?i)^Node is already registered with Rocket Pool - current deposit contract is at 0x[0-9a-fA-F]{40}$", "Node already registered message incorrect"},
    }); err != nil {
        t.Fatal(err)
    } else {
        for _, msg := range messages { t.Error(msg) }
    }

}

