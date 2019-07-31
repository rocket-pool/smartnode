package passwords

import (
    "io"
    "io/ioutil"
    "testing"
)


// Test password manager functionality
func TestPasswordManager(t *testing.T) {

    // Create temporary password path
    passwordPath, err := ioutil.TempDir("", "")
    if err != nil { t.Fatal(err) }
    passwordPath += "/password"

    // Create temporary input file
    input, err := ioutil.TempFile("", "")
    if err != nil { t.Fatal(err) }
    defer input.Close()

    // Write input to file
    io.WriteString(input, "foobarbaz" + "\n")
    input.Seek(0, io.SeekStart)

    // Initialise password manager
    passwordManager := NewPasswordManager(input, passwordPath)

    // Check if password exists
    if passwordExists := passwordManager.PasswordExists(); passwordExists {
        t.Errorf("Incorrect password exists status: expected %t, got %t", false, passwordExists)
    }

    // Get passphrase
    if _, err := passwordManager.GetPassphrase(); err == nil {
        t.Error("Password manager GetPassphrase() method should return error when uninitialised")
    }

    // Create password
    if password, err := passwordManager.CreatePassword(); err != nil {
        t.Error(err)
    } else if password != "foobarbaz" {
        t.Errorf("Incorrect created password: expected %s, got %s", "foobarbaz", password)
    }

    // Overwrite password
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

