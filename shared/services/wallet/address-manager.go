package wallet

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/common"
)

var emptyAddress = common.Address{}

// File signature for storing the node address on disk
type WalletAddressFile struct {
	Address common.Address `json:"address"`
}

// Wallet address manager
type WalletAddressManager struct {
	path    string
	address common.Address
}

func NewWalletAddressManager(path string) *WalletAddressManager {
	return &WalletAddressManager{
		path: path,
	}
}

// Initialize the node address from the stored file on disk;
// returns true if it was initialized successfully or false if it wasn't (i.e. if it's not stored on disk)
func (wm *WalletAddressManager) InitAddress() (bool, error) {
	// Done if it's already initialized
	if wm.address != emptyAddress {
		return true, nil
	}

	// Check if the wallet address file exists on disk
	_, err := os.Stat(wm.path)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("error checking wallet address file path: %w", err)
	}

	// Load the wallet address file if it exists
	bytes, err := os.ReadFile(wm.path)
	if err != nil {
		return false, fmt.Errorf("error reading wallet address file: %w", err)
	}

	// Deserialize
	var file WalletAddressFile
	err = json.Unmarshal(bytes, &file)
	if err != nil {
		return false, fmt.Errorf("error deserializing wallet address file: %w", err)
	}

	wm.address = file.Address
	return true, nil
}

// Get the wallet address - if it isn't loaded yet, initialize it first
func (wm *WalletAddressManager) GetAddress() (common.Address, bool, error) {
	// Done if it's already initialized
	if wm.address != emptyAddress {
		return wm.address, true, nil
	}

	// Init and return the result
	isLoaded, err := wm.InitAddress()
	return wm.address, isLoaded, err
}

// Store the wallet address on disk
func (wm *WalletAddressManager) StoreAddress(address common.Address) error {
	// Check if the wallet address file exists on disk
	_, err := os.Stat(wm.path)
	if !os.IsNotExist(err) {
		return fmt.Errorf("wallet address is already set")
	}

	// Serialize it
	file := WalletAddressFile{
		Address: wm.address,
	}
	bytes, err := json.Marshal(file)
	if err != nil {
		return fmt.Errorf("error serializing wallet address file: %w", err)
	}

	// Write to disk
	if err := os.WriteFile(wm.path, bytes, FileMode); err != nil {
		return fmt.Errorf("error writing wallet address to disk: %w", err)
	}
	return nil
}

// Delete the address from disk
func (wm *WalletAddressManager) DeleteAddress() error {
	// Check if it exists
	_, err := os.Stat(wm.path)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("error checking wallet address file path: %w", err)
	}

	// Delete it
	err = os.Remove(wm.path)
	if err != nil {
		return fmt.Errorf("error deleting wallet address file: %w", err)
	}
	return nil
}
