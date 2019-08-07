package deposit

import (
    "io/ioutil"
    "testing"

    "github.com/rocket-pool/smartnode/shared/utils/eth"

    test "github.com/rocket-pool/smartnode/tests/utils"
    testapp "github.com/rocket-pool/smartnode/tests/utils/app"
)


// Test deposit reserve command
func TestDepositReserve(t *testing.T) {

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
    reserveArgs := testapp.GetAppArgs(dataPath, "", output.Name())
    initArgs := testapp.GetAppArgs(dataPath, initInput.Name(), "")
    registerArgs := testapp.GetAppArgs(dataPath, registerInput.Name(), "")
    appOptions := testapp.GetAppOptions(dataPath)

    // Attempt to reserve deposit for uninitialised node
    if err := app.Run(append(reserveArgs, "deposit", "reserve", "3m")); err == nil { t.Error("Should return error for uninitialised node") }

    // Initialise node
    if err := app.Run(append(initArgs, "node", "init")); err != nil { t.Fatal(err) }

    // Attempt to reserve deposit for unregistered node
    if err := app.Run(append(reserveArgs, "deposit", "reserve", "3m")); err == nil { t.Error("Should return error for unregistered node") }

    // Seed node account & register node
    if err := testapp.AppSeedNodeAccount(appOptions, eth.EthToWei(10)); err != nil { t.Fatal(err) }
    if err := app.Run(append(registerArgs, "node", "register")); err != nil { t.Fatal(err) }

    // Reserve deposit
    if err := app.Run(append(reserveArgs, "deposit", "reserve", "3m")); err != nil { t.Error(err) }

    // Reserve deposit with current deposit reservation
    if err := app.Run(append(reserveArgs, "deposit", "reserve", "3m")); err != nil { t.Error(err) }

    // Check output
    if messages, err := testapp.CheckOutput(output.Name(), []string{}, map[int][]string{
        1: []string{"(?i)^Making deposit reservation...$", "Making deposit reservation message incorrect"},
        2: []string{"(?i)^Deposit reservation made successfully, requiring \\d+\\.\\d+ ETH and \\d+\\.\\d+ RPL, with a staking duration of \\S+ and expiring at", "Deposit reservation made message incorrect"},
        3: []string{"(?i)^Node has a current deposit reservation, please cancel or complete it$", "Deposit reservation exists message incorrect"},
    }); err != nil {
        t.Fatal(err)
    } else {
        for _, msg := range messages { t.Error(msg) }
    }

}

