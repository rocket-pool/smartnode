package validators

import (
    "bytes"
    "io/ioutil"
    "testing"

    "github.com/rocket-pool/smartnode/shared/services/validators"

    test "github.com/rocket-pool/smartnode/tests/utils"
)


// Test key manager functionality
func TestKeyManager(t *testing.T) {

    // Create temporary keychain path
    keychainPath, err := ioutil.TempDir("", "")
    if err != nil { t.Fatal(err) }
    keychainPath += "/keychain"

    // Create temporary password input file
    input, err := test.NewInputFile("foobarbaz" + "\n")
    if err != nil { t.Fatal(err) }
    defer input.Close()

    // Create password manager
    passwordManager, err := test.NewPasswordManager(input)
    if err != nil { t.Fatal(err) }

    // Initialise key manager
    keyManager := validators.NewKeyManager(keychainPath, passwordManager)

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
    } else if !bytes.Equal(createdKey.SecretKey.Marshal(), key.SecretKey.Marshal()) {
        t.Error("Incorrect retrieved validator key")
    }

    // Attempt to get nonexistent validator key
    if _, err := keyManager.GetValidatorKey(make([]byte, 48)); err == nil {
        t.Error("Retrieved nonexistent validator key")
    }

}

