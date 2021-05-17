package wallet

import (
	"bytes"
	"errors"
	"fmt"
	"sync"

	rptypes "github.com/rocket-pool/rocketpool-go/types"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
	eth2util "github.com/wealdtech/go-eth2-util"
)

// Config
const (
    ValidatorKeyPath = "m/12381/3600/%d/0/0"
    MaxValidatorKeyRecoverAttempts = 100
)


// Get the number of validator keys recorded in the wallet
func (w *Wallet) GetValidatorKeyCount() (uint, error) {

    // Check wallet is initialized
    if !w.IsInitialized() {
        return 0, errors.New("Wallet is not initialized")
    }

    // Return validator key count
    return w.ws.NextAccount, nil

}


// Get a validator key by index
func (w *Wallet) GetValidatorKeyAt(index uint) (*eth2types.BLSPrivateKey, error) {

    // Check wallet is initialized
    if !w.IsInitialized() {
        return nil, errors.New("Wallet is not initialized")
    }

    // Return validator key
    key, _, err := w.getValidatorPrivateKey(index)
    return key, err

}


// Get a validator key by public key
func (w *Wallet) GetValidatorKeyByPubkey(pubkey rptypes.ValidatorPubkey) (*eth2types.BLSPrivateKey, error) {

    // Check wallet is initialized
    if !w.IsInitialized() {
        return nil, errors.New("Wallet is not initialized")
    }

    // Get pubkey hex string
    pubkeyHex := pubkey.Hex()

    // Check for cached validator key index
    if index, ok := w.validatorKeyIndices[pubkeyHex]; ok {
        if key, _, err := w.getValidatorPrivateKey(index); err != nil {
            return nil, err
        } else if bytes.Equal(pubkey.Bytes(), key.PublicKey().Marshal()) {
            return key, nil
        }
    }

    // Find matching validator key
    var index uint
    var validatorKey *eth2types.BLSPrivateKey
    for index = 0; index < w.ws.NextAccount; index++ {
        if key, _, err := w.getValidatorPrivateKey(index); err != nil {
            return nil, err
        } else if bytes.Equal(pubkey.Bytes(), key.PublicKey().Marshal()) {
            validatorKey = key
            break
        }
    }

    // Check validator key
    if validatorKey == nil {
        return nil, fmt.Errorf("Validator %s key not found", pubkeyHex)
    }

    // Cache validator key index
    w.validatorKeyIndices[pubkeyHex] = index

    // Return
    return validatorKey, nil

}


// Create a new validator key
func (w *Wallet) CreateValidatorKey() (*eth2types.BLSPrivateKey, error) {

    // Check wallet is initialized
    if !w.IsInitialized() {
        return nil, errors.New("Wallet is not initialized")
    }

    // Get & increment account index
    index := w.ws.NextAccount
    w.ws.NextAccount++

    // Get validator key
    key, path, err := w.getValidatorPrivateKey(index)
    if err != nil {
        return nil, err
    }

    // Update keystores
    for name := range w.keystores {
        // Update the keystore in the wallet - using an iterator variable only runs it on the local copy
        if err := w.keystores[name].StoreValidatorKey(key, path); err != nil {
            return nil, fmt.Errorf("Could not store %s validator key: %w", name, err)
        }
    }

    // Return validator key
    return key, nil

}


// Recover a validator key by public key
func (w *Wallet) RecoverValidatorKey(pubkey rptypes.ValidatorPubkey) error {

    // Check wallet is initialized
    if !w.IsInitialized() {
        return errors.New("Wallet is not initialized")
    }

    // Find matching validator key
    var index uint
    var validatorKey *eth2types.BLSPrivateKey
    var derivationPath string
    for index = 0; index < w.ws.NextAccount + MaxValidatorKeyRecoverAttempts; index++ {
        if key, path, err := w.getValidatorPrivateKey(index); err != nil {
            return err
        } else if bytes.Equal(pubkey.Bytes(), key.PublicKey().Marshal()) {
            validatorKey = key
            derivationPath = path
            break
        }
    }

    // Check validator key
    if validatorKey == nil {
        return fmt.Errorf("Validator %s key not found", pubkey.Hex())
    }

    // Update account index
    nextIndex := index + 1
    if nextIndex > w.ws.NextAccount {
        w.ws.NextAccount = nextIndex
    }

    // Update keystores
    for name := range w.keystores {
        // Update the keystore in the wallet - using an iterator variable only runs it on the local copy
        if err := w.keystores[name].StoreValidatorKey(validatorKey, derivationPath); err != nil {
            return fmt.Errorf("Could not store %s validator key: %w", name, err)
        }
    }

    // Return
    return nil

}


// Get a validator private key by index
func (w *Wallet) getValidatorPrivateKey(index uint) (*eth2types.BLSPrivateKey, string, error) {

    // Get derivation path
    derivationPath := fmt.Sprintf(ValidatorKeyPath, index)

    // Check for cached validator key
    if validatorKey, ok := w.validatorKeys[index]; ok {
        return validatorKey, derivationPath, nil
    }

    // Initialize BLS support
    if err := initializeBLS(); err != nil {
        return nil, "", fmt.Errorf("Could not initialize BLS library: %w", err)
    }

    // Get private key
    privateKey, err := eth2util.PrivateKeyFromSeedAndPath(w.seed, derivationPath)
    if err != nil {
        return nil, "", fmt.Errorf("Could not get validator %d private key: %w", index, err)
    }

    // Cache validator key
    w.validatorKeys[index] = privateKey

    // Return
    return privateKey, derivationPath, nil

}


// Initialize BLS support
var initBLS sync.Once
func initializeBLS() error {
    var err error
    initBLS.Do(func() {
        err = eth2types.InitBLS()
    })
    return err
}

