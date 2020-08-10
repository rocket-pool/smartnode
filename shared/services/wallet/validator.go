package wallet

import (
    "bytes"
    "encoding/hex"
    "errors"
    "fmt"
    "sync"

    eth2types "github.com/wealdtech/go-eth2-types/v2"
    eth2util "github.com/wealdtech/go-eth2-util"
)


// Config
const ValidatorKeyPath = "m/12381/3600/%d/0/0"


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

    // Check for cached validator key
    if privateKey, ok := w.validatorKeys[index]; ok {
        return privateKey, nil
    }

    // Initialize BLS support
    initializeBLS()

    // Get private key
    privateKey, err := eth2util.PrivateKeyFromSeedAndPath(w.seed, fmt.Sprintf(ValidatorKeyPath, index))
    if err != nil {
        return nil, err
    }

    // Cache validator key
    w.validatorKeys[index] = privateKey

    // Return
    return privateKey, nil

}


// Get a validator key by public key
func (w *Wallet) GetValidatorKeyByPubkey(pubkey *eth2types.BLSPublicKey) (*eth2types.BLSPrivateKey, error) {

    // Check wallet is initialized
    if !w.IsInitialized() {
        return nil, errors.New("Wallet is not initialized")
    }

    // Encode pubkey
    pubkeyBytes := pubkey.Marshal()
    pubkeyHex := hex.EncodeToString(pubkeyBytes)

    // Check for cached validator key index
    if index, ok := w.validatorKeyIndices[pubkeyHex]; ok {
        return w.GetValidatorKeyAt(index)
    }

    // Find matching validator key
    var index uint
    var validatorKey *eth2types.BLSPrivateKey
    for index = 0; index < w.ws.NextAccount; index++ {
        key, err := w.GetValidatorKeyAt(index)
        if err != nil {
            return nil, err
        }
        if bytes.Equal(pubkeyBytes, key.PublicKey().Marshal()) {
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

    // Save wallet store
    if err := w.saveStore(); err != nil {
        return nil, err
    }

    // Return validator key
    return w.GetValidatorKeyAt(index)

}


// Initialize BLS support
var initBLS sync.Once
func initializeBLS() {
    initBLS.Do(func() {
        eth2types.InitBLS()
    })
}

