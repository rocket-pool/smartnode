package wallet

import (
	"github.com/rocket-pool/smartnode/bindings/types"
	rptypes "github.com/rocket-pool/smartnode/bindings/types"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
)

// Get the number of validator keys recorded in the wallet
func (w *masqueradeWallet) GetValidatorKeyCount() (uint, error) {
	return 0, ErrIsMasquerading

}

// Get a validator key by index
func (w *masqueradeWallet) GetValidatorKeyAt(index uint) (*eth2types.BLSPrivateKey, error) {
	return nil, ErrIsMasquerading

}

// Get a validator key by public key
func (w *masqueradeWallet) GetValidatorKeyByPubkey(pubkey rptypes.ValidatorPubkey) (*eth2types.BLSPrivateKey, error) {
	return nil, ErrIsMasquerading

}

// Create a new validator key
func (w *masqueradeWallet) CreateValidatorKey() (*eth2types.BLSPrivateKey, error) {
	return nil, ErrIsMasquerading

}

// Stores a validator key into all of the wallet's keystores
func (w *masqueradeWallet) StoreValidatorKey(key *eth2types.BLSPrivateKey, path string) error {
	return ErrIsMasquerading
}

// Loads a validator key from the wallet's keystores
func (w *masqueradeWallet) LoadValidatorKey(pubkey types.ValidatorPubkey) (*eth2types.BLSPrivateKey, error) {
	return nil, ErrIsMasquerading

}

// Deletes all of the keystore directories and persistent VC storage
func (w *masqueradeWallet) DeleteValidatorStores() error {
	return ErrIsMasquerading

}

// Returns the next validator key that will be generated without saving it
func (w *masqueradeWallet) GetNextValidatorKey() (*eth2types.BLSPrivateKey, error) {
	return nil, ErrIsMasquerading

}

// Recover a set of validator keys by their public key
func (w *masqueradeWallet) GetValidatorKeys(startIndex uint, length uint) ([]ValidatorKey, error) {
	return nil, ErrIsMasquerading

}

// Save a validator key
func (w *masqueradeWallet) SaveValidatorKey(key ValidatorKey) error {
	return ErrIsMasquerading

}

// Recover a validator key by public key
func (w *masqueradeWallet) RecoverValidatorKey(pubkey rptypes.ValidatorPubkey, startIndex uint) (uint, error) {
	return 0, ErrIsMasquerading

}

// Test recovery of a validator key by public key
func (w *masqueradeWallet) TestRecoverValidatorKey(pubkey rptypes.ValidatorPubkey, startIndex uint) (uint, error) {
	return 0, ErrIsMasquerading

}
