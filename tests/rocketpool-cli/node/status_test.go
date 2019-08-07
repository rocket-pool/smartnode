package node

import (
    "io/ioutil"
    "testing"

    "github.com/rocket-pool/smartnode/shared/utils/eth"

    test "github.com/rocket-pool/smartnode/tests/utils"
    testapp "github.com/rocket-pool/smartnode/tests/utils/app"
)


// Test node status command
func TestNodeStatus(t *testing.T) {

    // Create test app
    app := testapp.NewApp()

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
    statusArgs := testapp.GetAppArgs(dataPath, "", output.Name())
    initArgs := testapp.GetAppArgs(dataPath, initInput.Name(), "")
    registerArgs := testapp.GetAppArgs(dataPath, registerInput.Name(), "")
    appOptions := testapp.GetAppOptions(dataPath)

    // Attempt to get status of uninitialised node
    if err := app.Run(append(statusArgs, "node", "status")); err == nil { t.Error("Should return error for uninitialised node") }

    // Initialise node
    if err := app.Run(append(initArgs, "node", "init")); err != nil { t.Fatal(err) }
    
    // Get status of unregistered node
    if err := app.Run(append(statusArgs, "node", "status")); err != nil { t.Error(err) }

    // Seed node account & register node
    if err := testapp.AppSeedNodeAccount(appOptions, eth.EthToWei(10), nil); err != nil { t.Fatal(err) }
    if err := app.Run(append(registerArgs, "node", "register")); err != nil { t.Fatal(err) }

    // Get status of registered node
    if err := app.Run(append(statusArgs, "node", "status")); err != nil { t.Error(err) }

    // Make node trusted
    if err := testapp.AppSetNodeTrusted(appOptions); err != nil { t.Fatal(err) }

    // Get status of trusted node
    if err := app.Run(append(statusArgs, "node", "status")); err != nil { t.Error(err) }

    // Check output
    if messages, err := testapp.CheckOutput(output.Name(), []string{}, map[int][]string{
        1: []string{"(?i)^Node account 0x[0-9a-fA-F]{40} has a balance of \\d+\\.\\d+ ETH, \\d+\\.\\d+ rETH and \\d+\\.\\d+ RPL$", "Node account message incorrect"},
        3: []string{"(?i)^Node account 0x[0-9a-fA-F]{40} has a balance of \\d+\\.\\d+ ETH, \\d+\\.\\d+ rETH and \\d+\\.\\d+ RPL$", "Node account message incorrect"},
        5: []string{"(?i)^Node account 0x[0-9a-fA-F]{40} has a balance of \\d+\\.\\d+ ETH, \\d+\\.\\d+ rETH and \\d+\\.\\d+ RPL$", "Node account message incorrect"},
        2: []string{"(?i)^Node is not registered with Rocket Pool$", "Node not registered message incorrect"},
        4: []string{"(?i)^Node registered with Rocket Pool with contract at 0x[0-9a-fA-F]{40}, timezone '\\w+/\\w+' and a balance of \\d+\\.\\d+ ETH and \\d+\\.\\d+ RPL$", "Node registered message incorrect"},
        6: []string{"(?i)^Node registered with Rocket Pool with contract at 0x[0-9a-fA-F]{40}, timezone '\\w+/\\w+' and a balance of \\d+\\.\\d+ ETH and \\d+\\.\\d+ RPL$", "Node registered message incorrect"},
        7: []string{"(?i)^Node is a trusted Rocket Pool node and will perform watchtower duties$", "Node trusted message incorrect"},
    }); err != nil {
        t.Fatal(err)
    } else {
        for _, msg := range messages { t.Error(msg) }
    }

}

