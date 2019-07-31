package accounts

import (
    "io"
    "io/ioutil"
    "testing"

    "github.com/rocket-pool/smartnode/shared/services/passwords"
)


// Test account manager functionality
func TestAccountManager(t *testing.T) {

    // Create temporary password & keychain path
    path, err := ioutil.TempDir("", "")
    if err != nil { t.Fatal(err) }
    passwordPath := path + "/password"
    keychainPath := path + "/keychain"

    // Create temporary input file
    input, err := ioutil.TempFile("", "")
    if err != nil { t.Fatal(err) }
    defer input.Close()

    // Write input to file
    io.WriteString(input, "foobarbaz" + "\n")
    input.Seek(0, io.SeekStart)

    // Initialise password manager & account manager
    passwordManager := passwords.NewPasswordManager(input, passwordPath)
    accountManager := NewAccountManager(keychainPath, passwordManager)

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
    if _, err := passwordManager.CreatePassword(); err != nil { t.Fatal(err) }

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

