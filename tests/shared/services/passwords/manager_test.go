package passwords

import (
    "io/ioutil"
    "testing"

    "github.com/rocket-pool/smartnode/shared/services/passwords"

    test "github.com/rocket-pool/smartnode/tests/utils"
)


// Test password manager functionality
func TestPasswordManager(t *testing.T) {

    // Create temporary password input file
    input, err := test.NewInputFile("foobarbaz" + "\n")
    if err != nil { t.Fatal(err) }
    defer input.Close()

    // Create temporary password path
    passwordPath, err := ioutil.TempDir("", "")
    if err != nil { t.Fatal(err) }
    passwordPath += "/password"

    // Initialise password manager
    passwordManager := passwords.NewPasswordManager(input, passwordPath)

    // Check if password exists
    if passwordExists := passwordManager.PasswordExists(); passwordExists {
        t.Errorf("Incorrect password exists status: expected %t, got %t", false, passwordExists)
    }

    // Attempt to get passphrase while uninitialised
    if _, err := passwordManager.GetPassphrase(); err == nil {
        t.Error("Password manager GetPassphrase() method should return error when uninitialised")
    }

    // Create password
    if password, err := passwordManager.CreatePassword(); err != nil {
        t.Error(err)
    } else if password != "foobarbaz" {
        t.Errorf("Incorrect created password: expected %s, got %s", "foobarbaz", password)
    }

    // Attempt to create password again
    if _, err := passwordManager.CreatePassword(); err == nil {
        t.Error("Password manager CreatePassword() method should return error when initialised")
    }

    // Check if password exists
    if passwordExists := passwordManager.PasswordExists(); !passwordExists {
        t.Errorf("Incorrect password exists status: expected %t, got %t", true, passwordExists)
    }

    // Get passphrase
    expectedPassphrase := "69a0dafe010dfa7ba062ea986bf94d20f16cf49e376e761bf679b6cc5b8cee6d"
    if passphrase, err := passwordManager.GetPassphrase(); err != nil {
        t.Error(err)
    } else if passphrase != expectedPassphrase {
        t.Errorf("Incorrect passphrase: expected %s, got %s", expectedPassphrase, passphrase)
    }

}

