package passwords

import (
    "errors"
    "fmt"
    "io/ioutil"
)


// Config
const MinPasswordLength = 8


// Password manager
type PasswordManager struct {
    passwordPath string
}


// Create new password manager
func NewPasswordManager(passwordPath string) *PasswordManager {
    return &PasswordManager{
        passwordPath: passwordPath,
    }
}


// Check if the stored password exists
func (pm *PasswordManager) PasswordExists() bool {
    _, err := ioutil.ReadFile(pm.passwordPath)
    return (err == nil)
}


// Get the stored password
func (pm *PasswordManager) GetPassword() (string, error) {

    // Read the stored password from disk
    password, err := ioutil.ReadFile(pm.passwordPath)
    if err != nil {
        return "", errors.New("Could not read password from disk")
    }

    // Return
    return password, nil

}


// Set the stored password
func (pm *PasswordManager) SetPassword(password string) error {

    // Check password is not set
    if pm.PasswordExists() {
        return errors.New("Password is already set")
    }

    // Check password length
    if len(password) < MinPasswordLength {
        return fmt.Errorf("Password must be at least %d characters long", MinPasswordLength)
    }

    // Write to file
    if err := ioutil.WriteFile(pm.passwordPath, []byte(password), 0600); err != nil {
        return errors.New("Could not write password to disk")
    } else {
        return nil
    }

}

