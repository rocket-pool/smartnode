package lighthouse

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "path/filepath"

    "github.com/google/uuid"
    rptypes "github.com/rocket-pool/rocketpool-go/types"
    eth2types "github.com/wealdtech/go-eth2-types/v2"
    eth2ks "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"

    "github.com/rocket-pool/smartnode/shared/services/passwords"
    hexutil "github.com/rocket-pool/smartnode/shared/utils/hex"
)


// Config
const (
    KeystoreDir = "lighthouse"
    KeyFileName = "voting-keystore.json"
    FileMode = 0600
)


// Lighthouse keystore
type Keystore struct {
    keystorePath string
    pm *passwords.PasswordManager
    encryptor *eth2ks.Encryptor
}


// Encrypted validator key store
type validatorKey struct {
    Crypto map[string]interface{}   `json:"crypto"`
    Version uint                    `json:"version"`
    UUID uuid.UUID                  `json:"uuid"`
    Path string                     `json:"path"`
    Pubkey rptypes.ValidatorPubkey  `json:"pubkey"`
}


// Create new lighthouse keystore
func NewKeystore(keystorePath string, passwordManager *passwords.PasswordManager) *Keystore {
    return &Keystore{
        keystorePath: keystorePath,
        pm: passwordManager,
        encryptor: eth2ks.New(),
    }
}


// Store a validator key
func (ks *Keystore) StoreValidatorKey(key *eth2types.BLSPrivateKey, derivationPath string) error {

    // Get validator pubkey
    pubkey := rptypes.BytesToValidatorPubkey(key.PublicKey().Marshal())

    // Get wallet password
    password, err := ks.pm.GetPassword()
    if err != nil {
        return fmt.Errorf("Could not get wallet password: %w", err)
    }

    // Encrypt key
    encryptedKey, err := ks.encryptor.Encrypt(key.Marshal(), password)
    if err != nil {
        return fmt.Errorf("Could not encrypt validator key: %w", err)
    }

    // Create key store
    keyStore := validatorKey{
        Crypto: encryptedKey,
        Version: ks.encryptor.Version(),
        UUID: uuid.New(),
        Path: derivationPath,
        Pubkey: pubkey,
    }

    // Encode key store
    keyStoreBytes, err := json.Marshal(keyStore)
    if err != nil {
        return fmt.Errorf("Could not encode validator key: %w", err)
    }

    // Write key store to disk
    keyPath := filepath.Join(ks.keystorePath, KeystoreDir, hexutil.AddPrefix(pubkey.Hex()), KeyFileName)
    if err := ioutil.WriteFile(keyPath, keyStoreBytes, FileMode); err != nil {
        return fmt.Errorf("Could not write validator key to disk: %w", err)
    }

    // Return
    return nil

}

