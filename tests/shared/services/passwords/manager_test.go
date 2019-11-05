package passwords

import (
    "io/ioutil"
    "testing"

    "github.com/rocket-pool/smartnode/shared/services/passwords"
)


// Test password manager functionality
func TestPasswordManager(t *testing.T) {

    // Create temporary password path
    passwordPath, err := ioutil.TempDir("", "")
    if err != nil { t.Fatal(err) }
    passwordPath += "/password"

    // Initialise password manager
    passwordManager := passwords.NewPasswordManager(passwordPath)

    // Check if password exists
    if passwordExists := passwordManager.PasswordExists(); passwordExists {
        t.Errorf("Incorrect password exists status: expected %t, got %t", false, passwordExists)
    }

    // Attempt to get passphrase while uninitialised
    if _, err := passwordManager.GetPassphrase(); err == nil {
        t.Error("Password manager GetPassphrase() method should return error when uninitialised")
    }

    // Create password
    if err := passwordManager.SetPassword("foobarbaz"); err != nil { t.Error(err) }

    // Attempt to create password again
    if err := passwordManager.SetPassword("foobarbaz"); err == nil {
        t.Error("Password manager SetPassword() method should return error when initialised")
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

