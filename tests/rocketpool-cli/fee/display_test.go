package fee

import (
    "io/ioutil"
    "testing"

    testapp "github.com/rocket-pool/smartnode/tests/utils/app"
)


// Test fee display command
func TestFeeDisplay(t *testing.T) {

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
    displayArgs := testapp.GetAppArgs(dataPath, "", output.Name())
    setArgs := testapp.GetAppArgs(dataPath, "", "")

    // Display fee before set
    if err := app.Run(append(displayArgs, "fee", "display")); err != nil { t.Error(err) }

    // Set fee
    if err := app.Run(append(setArgs, "fee", "set", "6.00")); err != nil { t.Fatal(err) }

    // Display fee after set
    if err := app.Run(append(displayArgs, "fee", "display")); err != nil { t.Error(err) }

    // Check output
    if messages, err := testapp.CheckOutput(output.Name(), []string{}, map[int][]string{
        1: []string{"(?i)^The current Rocket Pool user fee paid to node operators is \\d+\\.\\d+% of rewards$", "Current fee message incorrect"},
        3: []string{"(?i)^The current Rocket Pool user fee paid to node operators is \\d+\\.\\d+% of rewards$", "Current fee message incorrect"},
        2: []string{"(?i)^The target Rocket Pool user fee to vote for is not set$", "Target fee not set message incorrect"},
        4: []string{"(?i)^The target Rocket Pool user fee to vote for is \\d+\\.\\d+% of rewards$", "Target fee value message incorrect"},
    }); err != nil {
        t.Fatal(err)
    } else {
        for _, msg := range messages { t.Error(msg) }
    }

}

