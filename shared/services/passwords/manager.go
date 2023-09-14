package passwords

import (
	"fmt"
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
	password     []byte
}

// Create new password manager
func NewPasswordManager(passwordPath string) *PasswordManager {
	return &PasswordManager{
		passwordPath: passwordPath,
		password:     []byte{},
	}
}

// Initialize the password from the stored file on disk;
// returns true if it was initialized successfully or false if it wasn't (i.e. if it's not stored on disk)
func (pm *PasswordManager) InitPassword() (bool, error) {
	// Done if it's already initialized
	if len(pm.password) != 0 {
		return true, nil
	}

	// Check if the password file exists on disk
	_, err := os.Stat(pm.passwordPath)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("error checking password file path: %w", err)
	}

	// Load the password file if it exists
	password, err := os.ReadFile(pm.passwordPath)
	if err != nil {
		return false, fmt.Errorf("error reading password file: %w", err)
	}
	pm.password = password
	return true, nil
}

// Get the password - if it isn't loaded yet, initialize it first
func (pm *PasswordManager) GetPassword() ([]byte, bool, error) {
	// Done if it's already initialized
	if len(pm.password) != 0 {
		return pm.password, true, nil
	}

	// Init and return the result
	isLoaded, err := pm.InitPassword()
	return pm.password, isLoaded, err
}

// Store the password on disk
func (pm *PasswordManager) StorePassword(password []byte) error {
	// Check if the password file exists on disk
	_, err := os.Stat(pm.passwordPath)
	if !os.IsNotExist(err) {
		return fmt.Errorf("password is already set")
	}

	// Check password length
	if len(password) < MinPasswordLength {
		return fmt.Errorf("password must be at least %d characters long", MinPasswordLength)
	}

	// Write to disk
	if err := os.WriteFile(pm.passwordPath, []byte(password), FileMode); err != nil {
		return fmt.Errorf("error writing password to disk: %w", err)
	}
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
	if err != nil {
		return fmt.Errorf("error deleting password file: %w", err)
	}
	return nil
}
