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

    // Create test app context
    dataPath, err := ioutil.TempDir("", "")
    if err != nil { t.Fatal(err) }
    c := testapp.GetAppContext(dataPath)

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        PM: true,
        PasswordOptional: true,
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
        t.Error("Node password was not initialised successfully")
    }

    // Check node password cannot be initialised with existing password
    if canInitPassword := node.CanInitNodePassword(p); !canInitPassword.HadExistingPassword {
        t.Error("HadExistingPassword flag was not set with an existing password set")
    }

    // Attempt to initialise node password
    if _, err := node.InitNodePassword(p, "foobarbaz"); err == nil {
        t.Error("InitNodePassword() method did not return error with an existing password set")
    }

}


// Test node init account methods
func TestNodeInitAccount(t *testing.T) {

    // Create test app context
    dataPath, err := ioutil.TempDir("", "")
    if err != nil { t.Fatal(err) }
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

    // Check node account cannot be initialised without node password set
    if canInitAccount := node.CanInitNodeAccount(p); !canInitAccount.NodePasswordDidNotExist {
        t.Error("NodePasswordDidNotExist flag was not set without an existing node password")
    }

    // Attempt to initialise node account
    if _, err := node.InitNodeAccount(p); err == nil {
        t.Error("InitNodeAccount() method did not return error without an existing node password")
    }

    // Initialise node password
    if _, err := node.InitNodePassword(p, "foobarbaz"); err != nil { t.Fatal(err) }

    // Check node account can be initialised
    if canInitAccount := node.CanInitNodeAccount(p); canInitAccount.HadExistingAccount {
        t.Error("HadExistingAccount flag was set without an existing node account")
    }

    // Initialise node account
    if initAccount, err := node.InitNodeAccount(p); err != nil {
        t.Error(err)
    } else if !initAccount.Success {
        t.Error("Node account was not initialised successfully")
    }

    // Check node account cannot be initialised with existing account
    if canInitAccount := node.CanInitNodeAccount(p); !canInitAccount.HadExistingAccount {
        t.Error("HadExistingAccount flag was not set with an existing node account")
    }

    // Attempt to initialise node account
    if _, err := node.InitNodeAccount(p); err == nil {
        t.Error("InitNodeAccount() method did not return error with an existing node account")
    }

}

