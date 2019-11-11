package utils

import (
    "io/ioutil"

    "github.com/rocket-pool/smartnode/shared/services/passwords"
)


// Create a temporary password manager
func NewPasswordManager() (*passwords.PasswordManager, error) {

    // Create temporary password path
    passwordPath, err := ioutil.TempDir("", "")
    if err != nil { return nil, err }
    passwordPath += "/password"

    // Create and return password manager
    return passwords.NewPasswordManager(passwordPath), nil

}


// Create a temporary initialised password manager
func NewInitPasswordManager(password string) (*passwords.PasswordManager, error) {

    // Create password manager
    passwordManager, err := NewPasswordManager()
    if err != nil { return nil, err }

    // Initialise password
    if err := passwordManager.SetPassword(password); err != nil { return nil, err }

    // Return
    return passwordManager, nil

}

