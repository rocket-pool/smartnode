package minipool

import (
    "io/ioutil"
    "testing"

    "github.com/rocket-pool/smartnode/shared/utils/eth"

    test "github.com/rocket-pool/smartnode/tests/utils"
    testapp "github.com/rocket-pool/smartnode/tests/utils/app"
)


// Test minipool status command
func TestMinipoolStatus(t *testing.T) {

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

    // Attempt to get minipool status for uninitialised node
    if err := app.Run(append(statusArgs, "minipool", "status")); err == nil { t.Error("Should return error for uninitialised node") }

    // Initialise node, seed node account & register node
    if err := app.Run(append(initArgs, "node", "init")); err != nil { t.Fatal(err) }
    if err := testapp.AppSeedNodeAccount(appOptions, eth.EthToWei(5), nil); err != nil { t.Fatal(err) }
    if err := app.Run(append(registerArgs, "node", "register")); err != nil { t.Fatal(err) }

    // Get minipool status with no minipools
    if err := app.Run(append(statusArgs, "minipool", "status")); err != nil { t.Error(err) }

    // Create minipools
    minipoolAddresses, err := testapp.AppCreateNodeMinipools(appOptions, "3m", 3)
    if err != nil { t.Fatal(err) }

    // Get minipool status with minipools
    if err := app.Run(append(statusArgs, "minipool", "status")); err != nil { t.Error(err) }

    // Check output
    outputRules := map[int][]string{}
    outputRules[1] = []string{"(?i)^Node has 0 minipools$", "Minipool count message incorrect"}
    outputRules[2] = []string{"(?i)^Node has 3 minipools$", "Minipool count message incorrect"}
    for mi, address := range minipoolAddresses {
        outputRules[mi * 10 + 3]  = []string{"(?i)^---", "Minipool separator incorrect"}
        outputRules[mi * 10 + 4]  = []string{"(?i)^Address:\\s+" + address.Hex() + "$", "Minipool address output incorrect"}
        outputRules[mi * 10 + 5]  = []string{"(?i)^Status:\\s+\\w+$", "Minipool status output incorrect"}
        outputRules[mi * 10 + 6]  = []string{"(?i)^Status Updated Time:", "Minipool status time output incorrect"}
        outputRules[mi * 10 + 7]  = []string{"(?i)^Staking Duration:\\s+\\S+$", "Minipool staking duration output incorrect"}
        outputRules[mi * 10 + 8]  = []string{"(?i)^Node ETH Deposited:\\s+\\d+\\.\\d+$", "Minipool node ETH deposit output incorrect"}
        outputRules[mi * 10 + 9]  = []string{"(?i)^Node RPL Deposited:\\s+\\d+\\.\\d+$", "Minipool node RPL deposit output incorrect"}
        outputRules[mi * 10 + 10] = []string{"(?i)^Deposit Count:\\s+\\d+$", "Minipool deposit count output incorrect"}
        outputRules[mi * 10 + 11] = []string{"(?i)^User Deposit Capacity:\\s+\\d+\\.\\d+$", "Minipool user deposit capacity output incorrect"}
        outputRules[mi * 10 + 12] = []string{"(?i)^User Deposit Total:\\s+\\d+\\.\\d+$", "Minipool user deposit total output incorrect"}
    }
    if messages, err := testapp.CheckOutput(output.Name(), []string{}, outputRules); err != nil {
        t.Fatal(err)
    } else {
        for _, msg := range messages { t.Error(msg) }
    }

}

