package utils

import (
    "io/ioutil"

    "github.com/rocket-pool/smartnode/shared/services/passwords"
    "github.com/rocket-pool/smartnode/shared/services/validators"
)


// Create a temporary key manager
func NewKeyManager(passwordManager *passwords.PasswordManager) (*validators.KeyManager, error) {

    // Create temporary keychain path
    keychainPath, err := ioutil.TempDir("", "")
    if err != nil { return nil, err }
    keychainPath += "/keychain"

    // Create and return key manager
    return validators.NewKeyManager(keychainPath, passwordManager), nil

}


// Create a temporary initialised key manager
func NewInitKeyManager(password string) (*validators.KeyManager, error) {

    // Create initialised password manager
    passwordManager, err := NewInitPasswordManager(password)
    if err != nil { return nil, err }

    // Create and return key manager
    return NewKeyManager(passwordManager)

}

