package node

import (
    "bufio"
    "io/ioutil"
    "testing"
    "os"
    "regexp"

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
    args := test.GetAppArgs(dataPath, input.Name(), output.Name())

    // Initialise node
    if err := app.Run(append(args, "node", "init")); err != nil { t.Error(err) }

    // Attempt to initialise node again
    if err := app.Run(append(args, "node", "init")); err != nil { t.Error(err) }

    // Read & check output
    output, err = os.Open(output.Name())
    if err != nil { t.Fatal(err) }
    line := 0
    for scanner := bufio.NewScanner(output); scanner.Scan(); {
        switch line {
            case 0: if !regexp.MustCompile("(?i)^Please enter a node password").MatchString(scanner.Text()) { t.Error("Password prompt message incorrect") }
            case 1: if !regexp.MustCompile("(?i)^Node password set successfully").MatchString(scanner.Text()) { t.Error("Node password set message incorrect") }
            case 2: if !regexp.MustCompile("(?i)^Node account created successfully").MatchString(scanner.Text()) { t.Error("Node account created message incorrect") }
            case 3: if !regexp.MustCompile("(?i)^Node password already set").MatchString(scanner.Text()) { t.Error("Node password already set message incorrect") }
            case 4: if !regexp.MustCompile("(?i)^Node account already exists").MatchString(scanner.Text()) { t.Error("Node account already exists message incorrect") }
        }
        line++
    }

}

