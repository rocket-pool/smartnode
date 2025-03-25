package wallet

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/rocket-pool/smartnode/bindings/types"
	rptypes "github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/shared/utils/validator"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
	eth2util "github.com/wealdtech/go-eth2-util"
)

// Config
const (
	MaxValidatorKeyRecoverAttempts uint = 1000
)

// A validator private/public key pair
type ValidatorKey struct {
	PublicKey      types.ValidatorPubkey
	PrivateKey     *eth2types.BLSPrivateKey
	DerivationPath string
	WalletIndex    uint
}

// Get the number of validator keys recorded in the wallet
func (w *hdWallet) GetValidatorKeyCount() (uint, error) {

	// Check wallet is initialized
	if !w.IsInitialized() {
		return 0, errors.New("Wallet is not initialized")
	}

	// Return validator key count
	return w.ws.NextAccount, nil

}

// Get a validator key by index
func (w *hdWallet) GetValidatorKeyAt(index uint) (*eth2types.BLSPrivateKey, error) {

	// Check wallet is initialized
	if !w.IsInitialized() {
		return nil, errors.New("Wallet is not initialized")
	}

	// Return validator key
	key, _, err := w.getValidatorPrivateKey(index)
	return key, err

}

// Get a validator key by public key
func (w *hdWallet) GetValidatorKeyByPubkey(pubkey rptypes.ValidatorPubkey) (*eth2types.BLSPrivateKey, error) {

	// Check wallet is initialized
	if !w.IsInitialized() {
		return nil, errors.New("Wallet is not initialized")
	}

	// Load the key from the wallet's keystores
	return w.LoadValidatorKey(pubkey)

}

// Create a new validator key
func (w *hdWallet) CreateValidatorKey() (*eth2types.BLSPrivateKey, error) {

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
	err = w.StoreValidatorKey(key, path)
	if err != nil {
		return nil, err
	}

	// Return validator key
	return key, nil

}

// Stores a validator key into all of the wallet's keystores
func (w *hdWallet) StoreValidatorKey(key *eth2types.BLSPrivateKey, path string) error {

	for name := range w.keystores {
		// Update the keystore in the wallet - using an iterator variable only runs it on the local copy
		if err := w.keystores[name].StoreValidatorKey(key, path); err != nil {
			return fmt.Errorf("Could not store %s validator key: %w", name, err)
		}
	}

	// Return validator key
	return nil

}

// Loads a validator key from the wallet's keystores
func (w *hdWallet) LoadValidatorKey(pubkey types.ValidatorPubkey) (*eth2types.BLSPrivateKey, error) {

	errors := []string{}
	// Try loading the key from all of the keystores, caching errors but not breaking on them
	for name := range w.keystores {
		key, err := w.keystores[name].LoadValidatorKey(pubkey)
		if err != nil {
			errors = append(errors, err.Error())
		}
		if key != nil {
			return key, nil
		}
	}

	if len(errors) > 0 {
		// If there were errors, return them
		return nil, fmt.Errorf("encountered the following errors while trying to load the key for validator %s:\n%s", pubkey.Hex(), strings.Join(errors, "\n"))
	} else {
		// If there were no errors, the key just didn't exist
		return nil, fmt.Errorf("couldn't find the key for validator %s in any of the wallet's keystores", pubkey.Hex())
	}

}

// Deletes all of the keystore directories and persistent VC storage
func (w *hdWallet) DeleteValidatorStores() error {

	for name := range w.keystores {
		keystorePath := w.keystores[name].GetKeystoreDir()
		err := os.RemoveAll(keystorePath)
		if err != nil {
			return fmt.Errorf("error deleting validator directory for %s: %w", name, err)
		}
	}

	return nil

}

// Returns the next validator key that will be generated without saving it
func (w *hdWallet) GetNextValidatorKey() (*eth2types.BLSPrivateKey, error) {

	// Check wallet is initialized
	if !w.IsInitialized() {
		return nil, errors.New("Wallet is not initialized")
	}

	// Get account index
	index := w.ws.NextAccount

	// Get validator key
	key, _, err := w.getValidatorPrivateKey(index)
	if err != nil {
		return nil, err
	}

	// Return validator key
	return key, nil

}

// Recover a set of validator keys by their public key
func (w *hdWallet) GetValidatorKeys(startIndex uint, length uint) ([]ValidatorKey, error) {

	// Check wallet is initialized
	if !w.IsInitialized() {
		return nil, errors.New("Wallet is not initialized")
	}

	validatorKeys := make([]ValidatorKey, 0, length)
	for index := startIndex; index < startIndex+length; index++ {
		key, path, err := w.getValidatorPrivateKey(index)
		if err != nil {
			return nil, fmt.Errorf("error getting validator key for index %d: %w", index, err)
		}
		validatorKey := ValidatorKey{
			PublicKey:      types.BytesToValidatorPubkey(key.PublicKey().Marshal()),
			PrivateKey:     key,
			DerivationPath: path,
			WalletIndex:    index,
		}
		validatorKeys = append(validatorKeys, validatorKey)
	}

	return validatorKeys, nil

}

// Save a validator key
func (w *hdWallet) SaveValidatorKey(key ValidatorKey) error {

	// Update account index
	if key.WalletIndex >= w.ws.NextAccount {
		w.ws.NextAccount = key.WalletIndex + 1
	}

	// Update keystores
	for name := range w.keystores {
		// Update the keystore in the wallet - using an iterator variable only runs it on the local copy
		if err := w.keystores[name].StoreValidatorKey(key.PrivateKey, key.DerivationPath); err != nil {
			return fmt.Errorf("could not store validator key %s in %s keystore: %w", key.PublicKey.Hex(), name, err)
		}
	}

	// Return
	return nil

}

// Recover a validator key by public key
func (w *hdWallet) RecoverValidatorKey(pubkey rptypes.ValidatorPubkey, startIndex uint) (uint, error) {

	// Check wallet is initialized
	if !w.IsInitialized() {
		return 0, errors.New("Wallet is not initialized")
	}

	// Find matching validator key
	var index uint
	var validatorKey *eth2types.BLSPrivateKey
	var derivationPath string
	for index = 0; index < MaxValidatorKeyRecoverAttempts; index++ {
		if key, path, err := w.getValidatorPrivateKey(index + startIndex); err != nil {
			return 0, err
		} else if bytes.Equal(pubkey.Bytes(), key.PublicKey().Marshal()) {
			validatorKey = key
			derivationPath = path
			break
		}
	}

	// Check validator key
	if validatorKey == nil {
		return 0, fmt.Errorf("Validator %s key not found", pubkey.Hex())
	}

	// Update account index
	nextIndex := index + startIndex + 1
	if nextIndex > w.ws.NextAccount {
		w.ws.NextAccount = nextIndex
	}

	// Update keystores
	for name := range w.keystores {
		// Update the keystore in the wallet - using an iterator variable only runs it on the local copy
		if err := w.keystores[name].StoreValidatorKey(validatorKey, derivationPath); err != nil {
			return 0, fmt.Errorf("Could not store %s validator key: %w", name, err)
		}
	}

	// Return
	return index + startIndex, nil

}

// Test recovery of a validator key by public key
func (w *hdWallet) TestRecoverValidatorKey(pubkey rptypes.ValidatorPubkey, startIndex uint) (uint, error) {

	// Check wallet is initialized
	if !w.IsInitialized() {
		return 0, errors.New("Wallet is not initialized")
	}

	// Find matching validator key
	var index uint
	var validatorKey *eth2types.BLSPrivateKey
	for index = 0; index < MaxValidatorKeyRecoverAttempts; index++ {
		if key, _, err := w.getValidatorPrivateKey(index + startIndex); err != nil {
			return 0, err
		} else if bytes.Equal(pubkey.Bytes(), key.PublicKey().Marshal()) {
			validatorKey = key
			break
		}
	}

	// Check validator key
	if validatorKey == nil {
		return 0, fmt.Errorf("Validator %s key not found", pubkey.Hex())
	}

	// Return
	return index + startIndex, nil

}

// Get a validator private key by index
func (w *hdWallet) getValidatorPrivateKey(index uint) (*eth2types.BLSPrivateKey, string, error) {

	// Get derivation path
	derivationPath := fmt.Sprintf(validator.ValidatorKeyPath, index)

	// Check for cached validator key
	if validatorKey, ok := w.validatorKeys[index]; ok {
		return validatorKey, derivationPath, nil
	}

	// Initialize BLS support
	if err := validator.InitializeBLS(); err != nil {
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
