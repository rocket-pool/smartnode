package deposit

import (
    "io/ioutil"
    "testing"

    testapp "github.com/rocket-pool/smartnode/tests/utils/app"
)


// Test deposit required command
func TestDepositRequired(t *testing.T) {

    // Create test app
    app := testapp.NewApp()

    // Create temporary output file
    output, err := ioutil.TempFile("", "")
    if err != nil { t.Fatal(err) }
    output.Close()

    // Create temporary data path
    dataPath, err := ioutil.TempDir("", "")
    if err != nil { t.Fatal(err) }

    // Get app args
    args := testapp.GetAppArgs(dataPath, "", output.Name())

    // Get required RPL
    if err := app.Run(append(args, "deposit", "required", "3m")); err != nil { t.Error(err) }

    // Check output
    if messages, err := testapp.CheckOutput(output.Name(), []string{}, map[int][]string{
        1: []string{"(?i)^\\d+\\.\\d+ RPL required to cover a deposit amount of \\d+\\.\\d+ ETH for \\S+ @ \\d+\\.\\d+ RPL / ETH$", "RPL required message incorrect"},
    }); err != nil {
        t.Fatal(err)
    } else {
        for _, msg := range messages { t.Error(msg) }
    }

}

