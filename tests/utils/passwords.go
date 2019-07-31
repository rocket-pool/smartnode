package utils

import (
    "io/ioutil"
    "os"

    "github.com/rocket-pool/smartnode/shared/services/passwords"
)


// Create a temporary password manager
func NewPasswordManager(input *os.File) (*passwords.PasswordManager, error) {

    // Create temporary password path
    passwordPath, err := ioutil.TempDir("", "")
    if err != nil { return nil, err }
    passwordPath += "/password"

    // Create and return password manager
    return passwords.NewPasswordManager(input, passwordPath), nil

}


// Create a temporary initialised password manager
func NewInitPasswordManager(password string) (*passwords.PasswordManager, error) {

    // Create password input file
    input, err := NewInputFile(password + "\n")
    if err != nil { return nil, err }
    defer input.Close()

    // Create password manager
    passwordManager, err := NewPasswordManager(input)
    if err != nil { return nil, err }

    // Initialise password
    if _, err := passwordManager.CreatePassword(); err != nil { return nil, err }

    // Return
    return passwordManager, nil

}

