package node

import (
    "io/ioutil"
    "testing"

    "github.com/rocket-pool/smartnode/shared/api/node"
    "github.com/rocket-pool/smartnode/shared/services"

    testapp "github.com/rocket-pool/smartnode/tests/utils/app"
)


// Test node init password methods
func TestNodeInitPassword(t *testing.T) {

    // Create temporary data path
    dataPath, err := ioutil.TempDir("", "")
    if err != nil { t.Fatal(err) }

    // Get test app context
    c := testapp.GetAppContext(dataPath)

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        PM: true,
        AM: true,
        PasswordOptional: true,
        NodeAccountOptional: true,
    })
    if err != nil { t.Fatal(err) }
    defer p.Cleanup()

    // Check node password can be initialised
    if canInitPassword := node.CanInitNodePassword(p); canInitPassword.HadExistingPassword {
        t.Error("HadExistingPassword flag was set without an existing password set")
    }

    // Initialise node password
    if initPassword, err := node.InitNodePassword(p, "foobarbaz"); err != nil {
        t.Error(err)
    } else if !initPassword.Success {
        t.Error("Node password was not set successfully")
    }

    // Check node password can be initialised
    if canInitPassword := node.CanInitNodePassword(p); !canInitPassword.HadExistingPassword {
        t.Error("HadExistingPassword flag was not set with an existing password set")
    }

    // Attempt to initialise node password
    if _, err := node.InitNodePassword(p, "foobarbaz"); err == nil {
        t.Error("InitNodePassword() method did not return error with an existing password set")
    }

}

