package node

import (
    "io/ioutil"
    "testing"

    test "github.com/rocket-pool/smartnode/tests/utils"
    testapp "github.com/rocket-pool/smartnode/tests/utils/app"
)


// Test node init command
func TestNodeInit(t *testing.T) {

    // Create test app
    app := testapp.NewApp()

    // Create temporary input file
    input, err := test.NewInputFile("foobarbaz" + "\n")
    if err != nil { t.Fatal(err) }
    input.Close()

    // Create temporary output file
    output, err := ioutil.TempFile("", "")
    if err != nil { t.Fatal(err) }
    output.Close()

    // Create temporary data path
    dataPath, err := ioutil.TempDir("", "")
    if err != nil { t.Fatal(err) }

    // Get app args
    args := testapp.GetAppArgs(dataPath, input.Name(), output.Name())

    // Initialise node
    if err := app.Run(append(args, "node", "init")); err != nil { t.Error(err) }

    // Attempt to initialise node again
    if err := app.Run(append(args, "node", "init")); err != nil { t.Error(err) }

    // Check output
    if messages, err := testapp.CheckOutput(output.Name(), []string{}, map[int][]string{
        1: []string{"(?i)^Please enter a node password", "Password prompt message incorrect"},
        2: []string{"(?i)^Node password set successfully: .{8,}$", "Node password set message incorrect"},
        3: []string{"(?i)^Node account created successfully: 0x[0-9a-fA-F]{40}$", "Node account created message incorrect"},
        4: []string{"(?i)^Node password already set.$", "Node password already set message incorrect"},
        5: []string{"(?i)^Node account already exists: 0x[0-9a-fA-F]{40}$", "Node account already exists message incorrect"},
    }); err != nil {
        t.Fatal(err)
    } else {
        for _, msg := range messages { t.Error(msg) }
    }

}

