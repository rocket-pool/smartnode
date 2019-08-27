package deposit

import (
    "io/ioutil"
    "math/big"
    "testing"

    "github.com/rocket-pool/smartnode/shared/utils/eth"

    test "github.com/rocket-pool/smartnode/tests/utils"
    testapp "github.com/rocket-pool/smartnode/tests/utils/app"
)


// Test deposit complete command
func TestDepositComplete(t *testing.T) {

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
    completeInput, err := test.NewInputFile(
        "YES" + "\n" +
        "YES" + "\n")
    if err != nil { t.Fatal(err) }
    completeInput.Close()

    // Create temporary output file
    output, err := ioutil.TempFile("", "")
    if err != nil { t.Fatal(err) }
    output.Close()

    // Create temporary data path
    dataPath, err := ioutil.TempDir("", "")
    if err != nil { t.Fatal(err) }

    // Get app args & options
    completeArgs := testapp.GetAppArgs(dataPath, completeInput.Name(), output.Name())
    initArgs := testapp.GetAppArgs(dataPath, initInput.Name(), "")
    registerArgs := testapp.GetAppArgs(dataPath, registerInput.Name(), "")
    reserveArgs := testapp.GetAppArgs(dataPath, "", "")
    appOptions := testapp.GetAppOptions(dataPath)

    // -- Uninitialised & unregistered

    // Attempt to complete deposit for uninitialised node
    if err := app.Run(append(completeArgs, "deposit", "complete")); err == nil { t.Error("Should return error for uninitialised node") }

    // Initialise node
    if err := app.Run(append(initArgs, "node", "init")); err != nil { t.Fatal(err) }

    // Attempt to complete deposit for unregistered node
    if err := app.Run(append(completeArgs, "deposit", "complete")); err == nil { t.Error("Should return error for unregistered node") }

    // Seed node account & register node
    if err := testapp.AppSeedNodeAccount(appOptions, eth.EthToWei(5), nil); err != nil { t.Fatal(err) }
    if err := app.Run(append(registerArgs, "node", "register")); err != nil { t.Fatal(err) }

    // -- No deposit reservation

    // Complete deposit without reservation
    if err := app.Run(append(completeArgs, "deposit", "complete")); err != nil { t.Error(err) }

    // -- Initial deposit (to set RPL ratio)

    // Make deposit reservation
    if err := app.Run(append(reserveArgs, "deposit", "reserve", "6m")); err != nil { t.Fatal(err) }

    // Get & check required deposit balances
    required, err := testapp.AppGetNodeRequiredBalances(appOptions)
    if err != nil { t.Fatal(err) }
    if required.EtherWei.Cmp(big.NewInt(0)) == 0 { t.Fatal("Required ETH should be > 0") }

    // Seed node contract
    if err := testapp.AppSeedNodeContract(appOptions, required.EtherWei, required.RplWei); err != nil { t.Fatal(err) }

    // Complete deposit
    if err := app.Run(append(completeArgs, "deposit", "complete")); err != nil { t.Error(err) }

    // -- Pay ETH & RPL from node account

    // Make deposit reservation
    if err := app.Run(append(reserveArgs, "deposit", "reserve", "6m")); err != nil { t.Fatal(err) }

    // Get & check required deposit balances
    required, err = testapp.AppGetNodeRequiredBalances(appOptions)
    if err != nil { t.Fatal(err) }
    if required.EtherWei.Cmp(big.NewInt(0)) == 0 { t.Fatal("Required ETH should be > 0") }
    if required.RplWei.Cmp(big.NewInt(0)) == 0 { t.Fatal("Required RPL should be > 0") }

    // Complete deposit with insufficient ETH balance
    if err := app.Run(append(completeArgs, "deposit", "complete")); err != nil { t.Error(err) }

    // Seed node account with ETH
    if err := testapp.AppSeedNodeAccount(appOptions, required.EtherWei, nil); err != nil { t.Fatal(err) }

    // Complete deposit with insufficient RPL balance
    if err := app.Run(append(completeArgs, "deposit", "complete")); err != nil { t.Error(err) }

    // Seed node account with RPL
    if err := testapp.AppSeedNodeAccount(appOptions, nil, required.RplWei); err != nil { t.Fatal(err) }

    // Complete deposit
    if err := app.Run(append(completeArgs, "deposit", "complete")); err != nil { t.Error(err) }

    // -- Seed ETH & RPL to node contract

    // Make deposit reservation
    if err := app.Run(append(reserveArgs, "deposit", "reserve", "6m")); err != nil { t.Fatal(err) }

    // Get & check required deposit balances
    required, err = testapp.AppGetNodeRequiredBalances(appOptions)
    if err != nil { t.Fatal(err) }
    if required.EtherWei.Cmp(big.NewInt(0)) == 0 { t.Fatal("Required ETH should be > 0") }
    if required.RplWei.Cmp(big.NewInt(0)) == 0 { t.Fatal("Required RPL should be > 0") }

    // Seed node contract
    if err := testapp.AppSeedNodeContract(appOptions, required.EtherWei, required.RplWei); err != nil { t.Fatal(err) }

    // Complete deposit
    if err := app.Run(append(completeArgs, "deposit", "complete")); err != nil { t.Error(err) }

    // -- Tests complete

    // Check output
    if messages, err := testapp.CheckOutput(output.Name(), []string{"(?i)^Node contract requires", "(?i)^Transferring RPL to node contract", "(?i)^Completing deposit", "(?i)^Processing deposit queue"}, map[int][]string{
        1: []string{"(?i)^Node does not have a current deposit reservation", "No deposit reservation message incorrect"},
        2: []string{"(?i)^Deposit completed successfully, minipool created at 0x[0-9a-fA-F]{40}$", "Deposit completed message incorrect"},
        5: []string{"(?i)^Deposit completed successfully, minipool created at 0x[0-9a-fA-F]{40}$", "Deposit completed message incorrect"},
        6: []string{"(?i)^Deposit completed successfully, minipool created at 0x[0-9a-fA-F]{40}$", "Deposit completed message incorrect"},
        3: []string{"(?i)^Node balance of \\d+\\.\\d+ ETH plus account balance of \\d+\\.\\d+ ETH is not enough to cover requirement of \\d+\\.\\d+ ETH$", "Insufficient ETH balance message incorrect"},
        4: []string{"(?i)^Node balance of \\d+\\.\\d+ RPL plus account balance of \\d+\\.\\d+ RPL is not enough to cover requirement of \\d+\\.\\d+ RPL$", "Insufficient RPL balance message incorrect"},
    }); err != nil {
        t.Fatal(err)
    } else {
        for _, msg := range messages { t.Error(msg) }
    }

}

