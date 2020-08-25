package prysm

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"

    eth2types "github.com/wealdtech/go-eth2-types/v2"
)


// Config
const (
    KeystoreDir = "prysm"
    WalletDir = "derived"
    WalletFileName = "seed.encrypted.json"
    ConfigFileName = "keymanageropts.json"
    DirMode = 0700
    FileMode = 0600

    DerivedPathStructure = "m / purpose / coin_type / account_index / withdrawal_key / validating_key"
    DerivedEIPNumber = "EIP-2334"
)


// Prysm keystore
type Keystore struct {
    keystorePath string
}


// Prysm wallet config
type walletConfig struct {
    DerivedPathStructure string `json:"DerivedPathStructure"`
    DerivedEIPNumber string     `json:"DerivedEIPNumber"`
}


// Create new prysm keystore
func NewKeystore(keystorePath string) *Keystore {
    return &Keystore{
        keystorePath: keystorePath,
    }
}


// Store a wallet
func (ks *Keystore) StoreWallet(walletData []byte) error {

    // Create & encode wallet config
    configData, err := json.Marshal(walletConfig{
        DerivedPathStructure: DerivedPathStructure,
        DerivedEIPNumber: DerivedEIPNumber,
    })
    if err != nil {
        return fmt.Errorf("Could not encode wallet config: %w", err)
    }

    // Get file paths
    walletFilePath := filepath.Join(ks.keystorePath, KeystoreDir, WalletDir, WalletFileName)
    configFilePath := filepath.Join(ks.keystorePath, KeystoreDir, WalletDir, ConfigFileName)

    // Create wallet dir
    if err := os.MkdirAll(filepath.Dir(walletFilePath), DirMode); err != nil {
        return fmt.Errorf("Could not create wallet folder: %w", err)
    }

    // Write wallet to disk
    if err := ioutil.WriteFile(walletFilePath, walletData, FileMode); err != nil {
        return fmt.Errorf("Could not write wallet to disk: %w", err)
    }

    // Write wallet config to disk
    if err := ioutil.WriteFile(configFilePath, configData, FileMode); err != nil {
        return fmt.Errorf("Could not write wallet config to disk: %w", err)
    }

    // Return
    return nil

}


// Store a validator key
func (ks *Keystore) StoreValidatorKey(key *eth2types.BLSPrivateKey, derivationPath string) error {
    return nil
}

