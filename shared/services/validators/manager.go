package validators

import (
    "encoding/hex"
    "errors"

    "github.com/rocket-pool/smartnode/shared/services/passwords"
    "github.com/rocket-pool/smartnode/shared/utils/bls/keystore"
)


// Key manager
type KeyManager struct {
    ks keystore.Store
    pm *passwords.PasswordManager
}


/**
 * Create new key manager
 */
func NewKeyManager(keychainPath string, passwordManager *passwords.PasswordManager) *KeyManager {
    return &KeyManager{
        ks: keystore.NewKeystore(keychainPath),
        pm: passwordManager,
    }
}


/**
 * Get a validator key by public key bytes
 */
func (km *KeyManager) GetValidatorKey(pubkey []byte) (*keystore.Key, error) {

    // Get keystore passphrase
    passphrase, err := km.pm.GetPassphrase()
    if err != nil {
        return nil, errors.New("Error retrieving node keystore passphrase: " + err.Error())
    }

    // Get all stored validator keys
    keys, err := km.ks.GetStoredKeys(passphrase)
    if err != nil {
        return nil, errors.New("Error retrieving stored validator keys: " + err.Error())
    }

    // Encode pubkey to search for
    pubkeyHex := make([]byte, hex.EncodedLen(len(pubkey)))
    hex.Encode(pubkeyHex, pubkey)
    pubkeyStr := string(pubkeyHex)

    // Return key if found
    if key, ok := keys[pubkeyStr]; !ok {
        return nil, errors.New("Validator key not found")
    } else {
        return key, nil
    }

}


/**
 * Create a validator key
 */
func (km *KeyManager) CreateValidatorKey() (*keystore.Key, error) {

    // Get keystore passphrase
    passphrase, err := km.pm.GetPassphrase()
    if err != nil {
        return nil, errors.New("Error retrieving node keystore passphrase: " + err.Error())
    }

    // Create key
    key, err := km.ks.NewKey(passphrase)
    if err != nil {
        return nil, errors.New("Error creating validator key: " + err.Error())
    }

    // Return
    return key, nil

}

