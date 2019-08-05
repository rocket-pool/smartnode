package node

import (
    "io/ioutil"
    "testing"

    test "github.com/rocket-pool/smartnode/tests/utils"
)


// Test node init command
func TestInitNode(t *testing.T) {

    // Create test app
    app := test.NewApp()

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
    args := test.AppArgs(input.Name(), output.Name(), dataPath)

    // Initialise node
    if err := app.Run(append(args, "node", "init")); err != nil { t.Error(err) }

    // Attempt to initialise when already initialised
    if err := app.Run(append(args, "node", "init")); err != nil { t.Error(err) }

    t.Error("!")

}

