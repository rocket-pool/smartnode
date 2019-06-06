package passwords

import (
    "bytes"
    "encoding/hex"
    "errors"
    "io/ioutil"

    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/cli"
    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/eth"
)


// Password hash salt
// This is added to the stored password and hashed to generate the keystore passphrase.
// DO NOT change this value, as doing so will break existing keystores.
const PASSWORD_SALT string = "iRmrlkOXNzOcEf8Dy3HQTgNNc4HYAMMeft7axN6XngLVei49OPR08aCl8oymF2ZG"


// Password manager
type PasswordManager struct {
    passwordPath string
}


/**
 * Create new password manager
 */
func NewPasswordManager(passwordPath string) *PasswordManager {
    return &PasswordManager{
        passwordPath: passwordPath,
    }
}


/**
 * Get the passphrase based on the hash of the stored password and salt
 */
func (pm *PasswordManager) GetPassphrase() (string, error) {

    // Get the stored password from disk
    password, err := ioutil.ReadFile(pm.passwordPath)
    if err != nil {
        return "", errors.New("Could not read password from disk.")
    }

    // Hash the password with salt and encode as hex
    passwordHash := eth.KeccakBytes(bytes.Join([][]byte{password, []byte(PASSWORD_SALT)}, []byte{}))
    passwordHashHex := make([]byte, hex.EncodedLen(len(passwordHash[:])))
    hex.Encode(passwordHashHex, passwordHash[:])

    // Return
    return string(passwordHashHex), nil

}


/**
 * Check if the stored password exists
 */
func (pm *PasswordManager) PasswordExists() bool {
    _, err := ioutil.ReadFile(pm.passwordPath)
    return (err == nil)
}


/**
 * Create the stored password
 */
func (pm *PasswordManager) CreatePassword() (string, error) {

    // Prompt for password
    password := cli.Prompt("Please enter a node password (this will be saved locally and used to generate dynamic keystore passphrases):", "^.{8,}$", "Please enter a password with 8 or more characters")

    // Write to file
    if err := ioutil.WriteFile(pm.passwordPath, []byte(password), 0600); err != nil {
        return "", errors.New("Could not write password to disk.")
    } else {
        return password, nil
    }

}

