package utils

import (
    "io/ioutil"

    "github.com/rocket-pool/smartnode/shared/services/accounts"
    "github.com/rocket-pool/smartnode/shared/services/passwords"
)


// Create a temporary account manager
func NewAccountManager(passwordManager *passwords.PasswordManager) (*accounts.AccountManager, error) {

    // Create temporary keychain path
    keychainPath, err := ioutil.TempDir("", "")
    if err != nil { return nil, err }
    keychainPath += "/keychain"

    // Create and return account manager
    return accounts.NewAccountManager(keychainPath, passwordManager), nil

}


// Create a temporary initialised account manager
func NewInitAccountManager(password string) (*accounts.AccountManager, error) {

    // Create initialised password manager
    passwordManager, err := NewInitPasswordManager(password)
    if err != nil { return nil, err }

    // Create account manager
    accountManager, err := NewAccountManager(passwordManager)
    if err != nil { return nil, err }

    // Initialise account
    if _, err := accountManager.CreateNodeAccount(); err != nil { return nil, err }

    // Return
    return accountManager, nil

}

