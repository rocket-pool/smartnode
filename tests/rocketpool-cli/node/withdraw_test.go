package node

import (
    "io/ioutil"
    "testing"

    "github.com/rocket-pool/smartnode/shared/utils/eth"

    test "github.com/rocket-pool/smartnode/tests/utils"
    testapp "github.com/rocket-pool/smartnode/tests/utils/app"
)


// Test node withdraw command
func TestNodeWithdraw(t *testing.T) {

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
    withdrawArgs := testapp.GetAppArgs(dataPath, "", output.Name())
    initArgs := testapp.GetAppArgs(dataPath, initInput.Name(), "")
    registerArgs := testapp.GetAppArgs(dataPath, registerInput.Name(), "")
    appOptions := testapp.GetAppOptions(dataPath)

    // Attempt to withdraw from uninitialised node
    if err := app.Run(append(withdrawArgs, "node", "withdraw", "5", "ETH")); err == nil { t.Error("Should return error for uninitialised node") }

    // Initialise node
    if err := app.Run(append(initArgs, "node", "init")); err != nil { t.Fatal(err) }

    // Attempt to withdraw from unregistered node
    if err := app.Run(append(withdrawArgs, "node", "withdraw", "5", "ETH")); err == nil { t.Error("Should return error for unregistered node") }

    // Seed node account & register node
    if err := testapp.AppSeedNodeAccount(appOptions, eth.EthToWei(10), nil); err != nil { t.Fatal(err) }
    if err := app.Run(append(registerArgs, "node", "register")); err != nil { t.Fatal(err) }

    // Attempt to withdraw from node with no balance
    if err := app.Run(append(withdrawArgs, "node", "withdraw", "5", "ETH")); err != nil { t.Error(err) }
    if err := app.Run(append(withdrawArgs, "node", "withdraw", "5", "RPL")); err != nil { t.Error(err) }

    // Seed node contract
    if err := testapp.AppSeedNodeContract(appOptions, eth.EthToWei(10), eth.EthToWei(10)); err != nil { t.Fatal(err) }

    // Withdraw from node with balance
    if err := app.Run(append(withdrawArgs, "node", "withdraw", "5", "ETH")); err != nil { t.Error(err) }
    if err := app.Run(append(withdrawArgs, "node", "withdraw", "5", "RPL")); err != nil { t.Error(err) }

    // Check output
    if messages, err := testapp.CheckOutput(output.Name(), []string{"(?i)^Withdrawing from node contract...$"}, map[int][]string{
        1: []string{"(?i)^Withdrawal amount exceeds available balance on node contract$", "Insufficient balance message incorrect"},
        2: []string{"(?i)^Withdrawal amount exceeds available balance on node contract$", "Insufficient balance message incorrect"},
        3: []string{"(?i)^Successfully withdrew \\d+\\.\\d+ (ETH|RPL) from node contract to account$", "Withdrawn message incorrect"},
        4: []string{"(?i)^Successfully withdrew \\d+\\.\\d+ (ETH|RPL) from node contract to account$", "Withdrawn message incorrect"},
    }); err != nil {
        t.Fatal(err)
    } else {
        for _, msg := range messages { t.Error(msg) }
    }

}

