package validator

import (
    "bytes"
    "encoding/hex"
    "io/ioutil"
    "os"
    "path/filepath"

    "github.com/rocket-pool/smartnode/shared/utils/bls"
    hexutil "github.com/rocket-pool/smartnode/shared/utils/hex"
)


// Lighthouse keystore settings
const LIGHTHOUSE_KEYSTORE_PATH string = "lighthouse"
const LIGHTHOUSE_KEY_FILENAME string = "voting_keypair"


// Read Lighthouse keys from keystore
func readLighthouseKeys(keystorePath string) (map[string]*bls.Key, error) {

    // Get lighthouse keystore path
    path := filepath.Join(keystorePath, LIGHTHOUSE_KEYSTORE_PATH)

    // Load all key dirs
    keyDirs, err := ioutil.ReadDir(path)
    if err != nil {
        return nil, err
    }

    // Read keys
    keys := make(map[string]*bls.Key)
    for _, keyDir := range keyDirs {

        // Get key file path
        keyFilePath := filepath.Clean(filepath.Join(path, keyDir.Name(), LIGHTHOUSE_KEY_FILENAME))

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


// Write Lighthouse key to keystore
func writeLighthouseKey(keystorePath string, key *bls.Key) error {

    // Get key filename
    filename := filepath.Join(keystorePath, LIGHTHOUSE_KEYSTORE_PATH, hexutil.AddPrefix(hex.EncodeToString(key.PublicKey.Marshal())), LIGHTHOUSE_KEY_FILENAME)

    // Get key file contents (public key - 16 null bytes - private key)
    contents := bytes.Join([][]byte{key.PublicKey.Marshal(), key.SecretKey.Marshal()}, make([]byte, 16))

    // Create the keystore directory
    if err := os.MkdirAll(filepath.Dir(filename), 0700); err != nil {
        return err
    }

    // Write key file
    f, err := ioutil.TempFile(filepath.Dir(filename), "." + filepath.Base(filename) + ".tmp")
    if err != nil {
        return err
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
        return err
    }
    if err := f.Close(); err != nil {
        return err
    }
    if err := os.Rename(f.Name(), filename); err != nil {
        return err
    }

    // Return
    return nil

}

