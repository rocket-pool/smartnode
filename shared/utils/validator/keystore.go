package validator

import (
    "bytes"
    "encoding/hex"
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"

    "github.com/rocket-pool/smartnode/shared/utils/bls"
)


// Key filename
const KEY_FILENAME string = "voting_keypair"


// Prysm keystore wrapper
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

    // Load all key dirs
    keyDirs, err := ioutil.ReadDir(ks.path)
    if err != nil {
        return nil, err
    }

    // Read keys
    keys := make(map[string]*bls.Key)
    for _, keyDir := range keyDirs {

        // Get key file path
        keyFilePath := filepath.Clean(filepath.Join(ks.path, keyDir.Name(), KEY_FILENAME))

        // Read key file
        keyBytes, err := ioutil.ReadFile(keyFilePath)
        if err != nil {
            return nil, err
        }

        // Decode secret key
        skBytes := keyBytes[64:]
        sk, err := bls.SecretKeyFromBytes(skBytes)
        if err != nil {
            return nil, err
        }

        // Construct key
        key, err := bls.NewKeyFromBLS(sk)
        if err != nil {
            return nil, err
        }

        // Add to map
        keys[hex.EncodeToString(key.PublicKey.Marshal())] = key

    }

    // Return
    return keys, nil

}


// Create, store and return a new key
func (ks *Keystore) NewKey(password string) (*bls.Key, error) {

    // Create new key
    key, err := bls.NewKey()
    if err != nil {
        return nil, err
    }

    // Get key filename
    filename := fmt.Sprintf("%s/0x%s/%s", ks.path, hex.EncodeToString(key.PublicKey.Marshal()), KEY_FILENAME)

    // Get key file contents (public key - 16 null bytes - private key)
    contents := bytes.Join([][]byte{key.PublicKey.Marshal(), key.SecretKey.Marshal()}, make([]byte, 16))

    // Create the keystore directory
    if err := os.MkdirAll(filepath.Dir(filename), 0700); err != nil {
        return nil, err
    }

    // Write key file
    f, err := ioutil.TempFile(filepath.Dir(filename), "." + filepath.Base(filename) + ".tmp")
    if err != nil {
        return nil, err
    }
    if _, err := f.Write(contents); err != nil {
        newErr := f.Close()
        if newErr != nil {
            err = newErr
        }
        newErr = os.Remove(f.Name())
        if newErr != nil {
            err = newErr
        }
        return nil, err
    }
    if err := f.Close(); err != nil {
        return nil, err
    }
    if err := os.Rename(f.Name(), filename); err != nil {
        return nil, err
    }

    // Return
    return key, nil

}

