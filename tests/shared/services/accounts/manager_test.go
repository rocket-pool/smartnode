package accounts

import (
    "io/ioutil"
    "testing"

    "github.com/rocket-pool/smartnode/shared/services/accounts"

    test "github.com/rocket-pool/smartnode/tests/utils"
)


// Test account manager functionality
func TestAccountManager(t *testing.T) {

    // Create temporary keychain path
    keychainPath, err := ioutil.TempDir("", "")
    if err != nil { t.Fatal(err) }
    keychainPath += "/keychain"

    // Create password manager
    passwordManager, err := test.NewPasswordManager()
    if err != nil { t.Fatal(err) }

    // Initialise account manager
    accountManager := accounts.NewAccountManager(keychainPath, passwordManager)

    // Check if node account exists
    if nodeAccountExists := accountManager.NodeAccountExists(); nodeAccountExists {
        t.Errorf("Incorrect node account exists status: expected %t, got %t", false, nodeAccountExists)
    }

    // Attempt to get node account while uninitialised
    if _, err := accountManager.GetNodeAccount(); err == nil {
        t.Error("Account manager GetNodeAccount() method should return error when uninitialised")
    }

    // Attempt to get node account transactor while uninitialised
    if _, err := accountManager.GetNodeAccountTransactor(); err == nil {
        t.Error("Account manager GetNodeAccountTransactor() method should return error when uninitialised")
    }

    // Attempt to create node account while password is uninitialised
    if _, err := accountManager.CreateNodeAccount(); err == nil {
        t.Error("Account manager CreateNodeAccount() method should return error when password is uninitialised")
    }

    // Initialise password
    if err := passwordManager.SetPassword("foobarbaz"); err != nil { t.Fatal(err) }

    // Create node account
    if _, err := accountManager.CreateNodeAccount(); err != nil { t.Error(err) }

    // Attempt to create node account again
    if _, err := accountManager.CreateNodeAccount(); err == nil {
        t.Error("Account manager CreateNodeAccount() method should return error when initialised")
    }

    // Check if node account exists
    if nodeAccountExists := accountManager.NodeAccountExists(); !nodeAccountExists {
        t.Errorf("Incorrect node account exists status: expected %t, got %t", true, nodeAccountExists)
    }

    // Get node account
    if _, err := accountManager.GetNodeAccount(); err != nil { t.Error(err) }

    // Get node account transactor
    if _, err := accountManager.GetNodeAccountTransactor(); err != nil { t.Error(err) }

}

