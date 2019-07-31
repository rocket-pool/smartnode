package validators

import (
    "bytes"
    "io"
    "io/ioutil"
    "testing"

    "github.com/rocket-pool/smartnode/shared/services/passwords"
)


// Test key manager functionality
func TestKeyManager(t *testing.T) {

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

    // Initialise password manager & key manager
    passwordManager := passwords.NewPasswordManager(input, passwordPath)
    keyManager := NewKeyManager(keychainPath, passwordManager)

    // Attempt to create validator key while password is uninitialised
    if _, err := keyManager.CreateValidatorKey(); err == nil {
        t.Error("Key manager CreateValidatorKey() method should return error when password is uninitialised")
    }

    // Initialise password
    if _, err := passwordManager.CreatePassword(); err != nil { t.Fatal(err) }

    // Create validator key
    createdKey, err := keyManager.CreateValidatorKey()
    if err != nil { t.Error(err) }

    // Get validator key by pubkey
    if key, err := keyManager.GetValidatorKey(createdKey.PublicKey.Marshal()); err != nil {
        t.Error(err)
    } else if bytes.Compare(createdKey.SecretKey.Marshal(), key.SecretKey.Marshal()) != 0 {
        t.Error("Incorrect retrieved validator key")
    }

    // Attempt to get nonexistent validator key
    if _, err := keyManager.GetValidatorKey(make([]byte, 48)); err == nil {
        t.Error("Retrieved nonexistent validator key")
    }

}

