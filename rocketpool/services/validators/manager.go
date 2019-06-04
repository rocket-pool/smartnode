package validators

import (
    "errors"

    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/bls/keystore"
)


// Keystore passphrase
const PASSPHRASE string = ""


// Key manager
type KeyManager struct {
    ks keystore.Store
}


/**
 * Create new key manager
 */
func NewKeyManager(keychainPath string) *KeyManager {
    return &KeyManager{
        ks: keystore.NewKeystore(keychainPath),
    }
}


/**
 * Create a validator key
 */
func (km *KeyManager) CreateValidatorKey() (*keystore.Key, error) {
    key, err := km.ks.NewKey(PASSPHRASE)
    if err != nil {
        return nil, errors.New("Error creating validator key: " + err.Error())
    }
    return key, nil
}

