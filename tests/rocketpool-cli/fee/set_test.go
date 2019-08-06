package fee

import (
    "io/ioutil"
    "testing"

    testapp "github.com/rocket-pool/smartnode/tests/utils/app"
)


// Test fee set command
func TestFeeSet(t *testing.T) {

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

    // Set fee
    if err := app.Run(append(args, "fee", "set", "7.00")); err != nil { t.Error(err) }

    // Check output
    if messages, err := testapp.CheckOutput(output.Name(), []string{}, map[int][]string{
        1: []string{"(?i)^Target user fee to vote for successfully set to \\d+\\.\\d+%$", "Target fee set message incorrect"},
    }); err != nil {
        t.Fatal(err)
    } else {
        for _, msg := range messages { t.Error(msg) }
    }

}

