package deposit

import (
    "io/ioutil"
    "math/big"
    "testing"

    "github.com/rocket-pool/smartnode/shared/utils/eth"

    test "github.com/rocket-pool/smartnode/tests/utils"
    testapp "github.com/rocket-pool/smartnode/tests/utils/app"
)


// Test deposit make command
func TestDepositMake(t *testing.T) {

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
    makeExitInput, err := test.NewInputFile("3" + "\n")
    if err != nil { t.Fatal(err) }
    makeExitInput.Close()
    makeCancelInput, err := test.NewInputFile("2" + "\n")
    if err != nil { t.Fatal(err) }
    makeCancelInput.Close()
    makeCompleteInput, err := test.NewInputFile(
        "1" + "\n" +
        "YES" + "\n" +
        "YES" + "\n")
    if err != nil { t.Fatal(err) }
    makeCompleteInput.Close()

    // Create temporary output file
    output, err := ioutil.TempFile("", "")
    if err != nil { t.Fatal(err) }
    output.Close()

    // Create temporary data path
    dataPath, err := ioutil.TempDir("", "")
    if err != nil { t.Fatal(err) }

    // Get app args & options
    makeExitArgs := testapp.GetAppArgs(dataPath, makeExitInput.Name(), output.Name())
    makeCancelArgs := testapp.GetAppArgs(dataPath, makeCancelInput.Name(), output.Name())
    makeCompleteArgs := testapp.GetAppArgs(dataPath, makeCompleteInput.Name(), output.Name())
    initArgs := testapp.GetAppArgs(dataPath, initInput.Name(), "")
    registerArgs := testapp.GetAppArgs(dataPath, registerInput.Name(), "")
    appOptions := testapp.GetAppOptions(dataPath)

    // -- Uninitialised & unregistered

    // Attempt to make deposit for uninitialised node
    if err := app.Run(append(makeExitArgs, "deposit", "make", "3m")); err == nil { t.Error("Should return error for uninitialised node") }

    // Initialise node
    if err := app.Run(append(initArgs, "node", "init")); err != nil { t.Fatal(err) }

    // Attempt to make deposit for unregistered node
    if err := app.Run(append(makeExitArgs, "deposit", "make", "3m")); err == nil { t.Error("Should return error for unregistered node") }

    // Seed node account & register node
    if err := testapp.AppSeedNodeAccount(appOptions, eth.EthToWei(10), nil); err != nil { t.Fatal(err) }
    if err := app.Run(append(registerArgs, "node", "register")); err != nil { t.Fatal(err) }

    // -- Exit & cancel

    // Reserve deposit and exit
    if err := app.Run(append(makeExitArgs, "deposit", "make", "3m")); err != nil { t.Error(err) }

    // Cancel reserved deposit
    if err := app.Run(append(makeCancelArgs, "deposit", "make", "3m")); err != nil { t.Error(err) }

    // -- Initial deposit (to set RPL ratio)

    // Reserve deposit and exit
    if err := app.Run(append(makeExitArgs, "deposit", "make", "6m")); err != nil { t.Fatal(err) }

    // Get & check required deposit balances
    required, err := testapp.AppGetNodeRequiredBalances(appOptions)
    if err != nil { t.Fatal(err) }
    if required.EtherWei.Cmp(big.NewInt(0)) == 0 { t.Fatal("Required ETH should be > 0") }

    // Seed node contract
    if err := testapp.AppSeedNodeContract(appOptions, required.EtherWei, required.RplWei); err != nil { t.Fatal(err) }

    // Complete deposit
    if err := app.Run(append(makeCompleteArgs, "deposit", "make", "6m")); err != nil { t.Error(err) }

    // -- Pay ETH & RPL from node account

    // Reserve deposit and exit
    if err := app.Run(append(makeExitArgs, "deposit", "make", "6m")); err != nil { t.Fatal(err) }

    // Get & check required deposit balances
    required, err = testapp.AppGetNodeRequiredBalances(appOptions)
    if err != nil { t.Fatal(err) }
    if required.EtherWei.Cmp(big.NewInt(0)) == 0 { t.Fatal("Required ETH should be > 0") }
    if required.RplWei.Cmp(big.NewInt(0)) == 0 { t.Fatal("Required RPL should be > 0") }

    // Complete deposit with insufficient ETH balance
    if err := app.Run(append(makeCompleteArgs, "deposit", "make", "6m")); err != nil { t.Error(err) }

    // Seed node account with ETH
    if err := testapp.AppSeedNodeAccount(appOptions, required.EtherWei, nil); err != nil { t.Fatal(err) }

    // Complete deposit with insufficient RPL balance
    if err := app.Run(append(makeCompleteArgs, "deposit", "make", "6m")); err != nil { t.Error(err) }

    // Seed node account with RPL
    if err := testapp.AppSeedNodeAccount(appOptions, nil, required.RplWei); err != nil { t.Fatal(err) }

    // Complete deposit
    if err := app.Run(append(makeCompleteArgs, "deposit", "make", "6m")); err != nil { t.Error(err) }

    // -- Seed ETH & RPL to node contract

    // Reserve deposit and exit
    if err := app.Run(append(makeExitArgs, "deposit", "make", "6m")); err != nil { t.Fatal(err) }

    // Get & check required deposit balances
    required, err = testapp.AppGetNodeRequiredBalances(appOptions)
    if err != nil { t.Fatal(err) }
    if required.EtherWei.Cmp(big.NewInt(0)) == 0 { t.Fatal("Required ETH should be > 0") }
    if required.RplWei.Cmp(big.NewInt(0)) == 0 { t.Fatal("Required RPL should be > 0") }

    // Seed node contract
    if err := testapp.AppSeedNodeContract(appOptions, required.EtherWei, required.RplWei); err != nil { t.Fatal(err) }

    // Complete deposit
    if err := app.Run(append(makeCompleteArgs, "deposit", "make", "6m")); err != nil { t.Error(err) }

    // -- Tests complete

    // Check output
    if messages, err := testapp.CheckOutput(output.Name(), []string{
        "(?i)^Making deposit reservation",
        "(?i)^Canceling deposit reservation",
        "(?i)^Transferring RPL to node contract",
        "(?i)^Completing deposit",
        "(?i)^Processing deposit queue",
        "(?i)^Node already has a deposit reservation",
        "(?i)^Node deposit contract has a balance of",
        "(?i)^Node contract requires",
        "(?i)^Would you like to",
        "(?i)^\\d\\. Complete the deposit",
        "(?i)^\\d\\. Cancel the deposit",
        "(?i)^\\d\\. Finish later",
    }, map[int][]string{
        1:  []string{"(?i)^Deposit reservation made successfully, requiring \\d+\\.\\d+ ETH and \\d+\\.\\d+ RPL, with a staking duration of \\S+ and expiring at", "Deposit reservation made message incorrect"},
        3:  []string{"(?i)^Deposit reservation made successfully, requiring \\d+\\.\\d+ ETH and \\d+\\.\\d+ RPL, with a staking duration of \\S+ and expiring at", "Deposit reservation made message incorrect"},
        5:  []string{"(?i)^Deposit reservation made successfully, requiring \\d+\\.\\d+ ETH and \\d+\\.\\d+ RPL, with a staking duration of \\S+ and expiring at", "Deposit reservation made message incorrect"},
        9:  []string{"(?i)^Deposit reservation made successfully, requiring \\d+\\.\\d+ ETH and \\d+\\.\\d+ RPL, with a staking duration of \\S+ and expiring at", "Deposit reservation made message incorrect"},
        2:  []string{"(?i)^Deposit reservation cancelled successfully$", "Deposit reservation canceled message incorrect"},
        4:  []string{"(?i)^Deposit completed successfully, minipool created at 0x[0-9a-fA-F]{40}$", "Deposit completed message incorrect"},
        8:  []string{"(?i)^Deposit completed successfully, minipool created at 0x[0-9a-fA-F]{40}$", "Deposit completed message incorrect"},
        10: []string{"(?i)^Deposit completed successfully, minipool created at 0x[0-9a-fA-F]{40}$", "Deposit completed message incorrect"},
        6:  []string{"(?i)^Node balance of \\d+\\.\\d+ ETH plus account balance of \\d+\\.\\d+ ETH is not enough to cover requirement of \\d+\\.\\d+ ETH$", "Insufficient ETH balance message incorrect"},
        7:  []string{"(?i)^Node balance of \\d+\\.\\d+ RPL plus account balance of \\d+\\.\\d+ RPL is not enough to cover requirement of \\d+\\.\\d+ RPL$", "Insufficient RPL balance message incorrect"},
    }); err != nil {
        t.Fatal(err)
    } else {
        for _, msg := range messages { t.Error(msg) }
    }

}

