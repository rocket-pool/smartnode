package prysm

import (
    "encoding/hex"
    "path/filepath"

    "github.com/rocket-pool/smartnode/shared/utils/bls"
)


// Prysm keystore settings
const KEYSTORE_PATH string = "prysm"
const KEY_FILENAME_PREFIX string = "validatorprivatekey"


// Prysm keystore
type Keystore struct {
    path string
    ks bls.Store
}


// Create new keystore
func NewKeystore(directory string) *Keystore {
    return &Keystore{
        path: filepath.Join(directory, KEYSTORE_PATH),
        ks: bls.NewKeystore(filepath.Join(directory, KEYSTORE_PATH)),
    }
}


// Get keys from the keystore directory
func (ks *Keystore) GetStoredKeys() (map[string]*bls.Key, error) {
    return ks.ks.GetKeys(ks.path, "", "", false)
}


// Write a key to the keystore directory
func (ks *Keystore) StoreKey(key *bls.Key) error {
    filename := KEY_FILENAME_PREFIX + hex.EncodeToString(key.PublicKey.Marshal())[:12]
    return ks.ks.StoreKey(ks.ks.JoinPath(filename), key, "")
}

