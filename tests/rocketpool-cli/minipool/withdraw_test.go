package minipool

import (
    "io/ioutil"
    "testing"

    "github.com/rocket-pool/smartnode/shared/utils/eth"

    test "github.com/rocket-pool/smartnode/tests/utils"
    testapp "github.com/rocket-pool/smartnode/tests/utils/app"
)


// Test minipool withdraw command
func TestMinipoolWithdraw(t *testing.T) {

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
    withdrawInput, err := test.NewInputFile("A" + "\n")
    if err != nil { t.Fatal(err) }
    withdrawInput.Close()

    // Create temporary output file
    output, err := ioutil.TempFile("", "")
    if err != nil { t.Fatal(err) }
    output.Close()

    // Create temporary data path
    dataPath, err := ioutil.TempDir("", "")
    if err != nil { t.Fatal(err) }

    // Get app args & options
    withdrawArgs := testapp.GetAppArgs(dataPath, withdrawInput.Name(), output.Name())
    initArgs := testapp.GetAppArgs(dataPath, initInput.Name(), "")
    registerArgs := testapp.GetAppArgs(dataPath, registerInput.Name(), "")
    appOptions := testapp.GetAppOptions(dataPath)

    // Attempt to withdraw from uninitialised node's minipools
    if err := app.Run(append(withdrawArgs, "minipool", "withdraw")); err == nil { t.Error("Should return error for uninitialised node") }

    // Initialise node
    if err := app.Run(append(initArgs, "node", "init")); err != nil { t.Fatal(err) }

    // Attempt to withdraw from unregistered node's minipools
    if err := app.Run(append(withdrawArgs, "minipool", "withdraw")); err == nil { t.Error("Should return error for unregistered node") }

    // Seed node account & register node
    if err := testapp.AppSeedNodeAccount(appOptions, eth.EthToWei(5), nil); err != nil { t.Fatal(err) }
    if err := app.Run(append(registerArgs, "node", "register")); err != nil { t.Fatal(err) }

    // Create minipools
    minipoolAddresses, err := testapp.AppCreateNodeMinipools(appOptions, "12m", 3)
    if err != nil { t.Fatal(err) }

    // Withdraw with no withdrawn minipools
    if err := app.Run(append(withdrawArgs, "minipool", "withdraw")); err != nil { t.Error(err) }

    // Stake, logout and withdraw minipools
    if _, accessorAddress, err := testapp.AppCreateGroupAccessor(appOptions); err != nil {
        t.Fatal(err)
    } else {
        if err := testapp.AppStakeAllMinipools(appOptions, "12m", accessorAddress); err != nil { t.Fatal(err) }
        if err := testapp.AppSetNodeTrusted(appOptions); err != nil { t.Fatal(err) }
        if err := testapp.AppWithdrawMinipools(appOptions, minipoolAddresses, eth.EthToWei(40)); err != nil { t.Fatal(err) }
    }

    // Withdraw from minipools
    if err := app.Run(append(withdrawArgs, "minipool", "withdraw")); err != nil { t.Error(err) }

    // Check output
    outputRules := map[int][]string{}
    outputRules[1] = []string{"(?i)^No minipools are currently available for withdrawal$", "No withdrawable minipools message incorrect"}
    for mi, address := range minipoolAddresses {
        outputRules[mi + 2] = []string{"(?i)^Successfully withdrew deposit from minipool " + address.Hex() + "$", "Minipool deposit withdrawn message incorrect"}
    }
    if messages, err := testapp.CheckOutput(output.Name(), []string{"(?i)^Please select a minipool to withdraw from", "(?i)^\\d+:\\s+0x[0-9a-fA-F]{40}$", "(?i)^Withdrawing deposit from minipool 0x[0-9a-fA-F]{40}"}, outputRules); err != nil {
        t.Fatal(err)
    } else {
        for _, msg := range messages { t.Error(msg) }
    }

}

