package validator

import (
    "github.com/rocket-pool/smartnode/shared/utils/bls"
)


// Validator keystore
type Keystore struct {
    path string
}


// Create new keystore
func NewKeystore(directory string) *Keystore {
    return &Keystore{
        path: directory,
    }
}


// Get keys from the keystore directory
func (ks *Keystore) GetStoredKeys(password string) (map[string]*bls.Key, error) {
    return readLighthouseKeys(ks.path)
}


// Create, store and return a new key
func (ks *Keystore) NewKey(password string) (*bls.Key, error) {

    // Create new key
    key, err := bls.NewKey()
    if err != nil {
        return nil, err
    }

    // Write to client keystores
    if err := writeLighthouseKey(ks.path, key); err != nil { return nil, err }
    if err := writePrysmKey(ks.path, key); err != nil { return nil, err }

    // Return
    return key, nil

}

