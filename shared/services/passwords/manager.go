package passwords

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

// Config
const (
	MinPasswordLength = 12
	FileMode          = 0600
)

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

// Check if the password has been set
func (pm *PasswordManager) IsPasswordSet() bool {
	_, err := ioutil.ReadFile(pm.passwordPath)
	return (err == nil)
}

// Get the password
func (pm *PasswordManager) GetPassword() (string, error) {

	// Read from disk
	password, err := ioutil.ReadFile(pm.passwordPath)
	if err != nil {
		return "", fmt.Errorf("Could not read password from disk: %w", err)
	}

	// Return
	return string(password), nil

}

// Set the password
func (pm *PasswordManager) SetPassword(password string) error {

	// Check password is not set
	if pm.IsPasswordSet() {
		return errors.New("Password is already set")
	}

	// Check password length
	if len(password) < MinPasswordLength {
		return fmt.Errorf("Password must be at least %d characters long", MinPasswordLength)
	}

	// Write to disk
	if err := ioutil.WriteFile(pm.passwordPath, []byte(password), FileMode); err != nil {
		return fmt.Errorf("Could not write password to disk: %w", err)
	}

	// Return
	return nil

}

// Delete the password
func (pm *PasswordManager) DeletePassword() error {

	// Check if it exists
	_, err := os.Stat(pm.passwordPath)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("error checking password file path: %w", err)
	}

	// Delete it
	err = os.Remove(pm.passwordPath)
	return err

}
