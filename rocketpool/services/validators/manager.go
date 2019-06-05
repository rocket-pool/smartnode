package validators

import (
    "encoding/hex"
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
 * Get a validator key by public key bytes
 */
func (km *KeyManager) GetValidatorKey(pubkey []byte) (*keystore.Key, error) {

    // Get all stored validator keys
    keys, err := km.ks.GetStoredKeys(PASSPHRASE)
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
    key, err := km.ks.NewKey(PASSPHRASE)
    if err != nil {
        return nil, errors.New("Error creating validator key: " + err.Error())
    }
    return key, nil
}

