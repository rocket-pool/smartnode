package keystore

import (
    "crypto/rand"

    pks "github.com/prysmaticlabs/prysm/shared/keystore"
)


// Prysm keystore wrapper
type Keystore struct {
    path string
    ks pks.Store
}


// Create new keystore
func NewKeystore(directory string) *Keystore {
    return &Keystore{
        path: directory,
        ks: pks.NewKeystore(directory),
    }
}


// Get keys from the keystore directory
func (ks *Keystore) GetStoredKeys(password string) (map[string]*pks.Key, error) {
    return ks.ks.GetKeys(ks.path, "", password)
}


// Create, store and return a new key
func (ks *Keystore) NewKey(password string) (*pks.Key, error) {
    key, err := pks.NewKey(rand.Reader)
    if err != nil {
        return nil, err
    }
    if err := ks.ks.StoreKey(ks.ks.JoinPath(keyFileName(key.PublicKey)), key, password); err != nil {
        return nil, err
    }
    return key, nil
}

