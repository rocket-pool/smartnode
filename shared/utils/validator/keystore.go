package validator

import (
    "github.com/rocket-pool/smartnode/shared/utils/bls"
    "github.com/rocket-pool/smartnode/shared/utils/validator/lighthouse"
    "github.com/rocket-pool/smartnode/shared/utils/validator/prysm"
)


// Validator keystore
type Keystore struct {
    path string
    lighthouse *lighthouse.Keystore
    prysm *prysm.Keystore
}


// Create new keystore
func NewKeystore(directory string) *Keystore {
    return &Keystore{
        path: directory,
        lighthouse: lighthouse.NewKeystore(directory),
        prysm: prysm.NewKeystore(directory),
    }
}


// Get keys from the keystore directory
// TODO: encryption not implemented
func (ks *Keystore) GetStoredKeys(password string) (map[string]*bls.Key, error) {
    return ks.prysm.GetStoredKeys()
}


// Create, store and return a new key
// TODO: encryption not implemented
func (ks *Keystore) NewKey(password string) (*bls.Key, error) {

    // Create new key
    key, err := bls.NewKey()
    if err != nil {
        return nil, err
    }

    // Write to client keystores
    if err := ks.lighthouse.StoreKey(key); err != nil { return nil, err }
    if err := ks.prysm.StoreKey(key); err != nil { return nil, err }

    // Return
    return key, nil

}

