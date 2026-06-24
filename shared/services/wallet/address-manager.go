package wallet

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
)

const (
	addressFileMode fs.FileMode = 0664
)

type addressFile struct {
	Address string `json:"address"`
	Observe bool   `json:"observe"`
}

// Simple class to wrap the node's address file
type AddressManager struct {
	path    string
	address common.Address
	observe bool
}

// Creates a new address manager
func NewAddressManager(path string) *AddressManager {
	return &AddressManager{
		path: path,
	}
}

// Gets the address saved on disk. Returns empty address if the address file doesn't exist.
// Also loads the observe flag as a side effect; check IsObserve() after calling.
func (m *AddressManager) LoadAddress() (common.Address, error) {
	m.address = common.Address{}
	m.observe = false

	_, err := os.Stat(m.path)
	if errors.Is(err, fs.ErrNotExist) {
		return common.Address{}, nil
	} else if err != nil {
		return common.Address{}, fmt.Errorf("error checking if address file exists: %w", err)
	}

	bytes, err := os.ReadFile(m.path)
	if err != nil {
		return common.Address{}, fmt.Errorf("error loading address file [%s]: %w", m.path, err)
	}

	var af addressFile
	if err := json.Unmarshal(bytes, &af); err != nil {
		// backward compat: plain hex address written by older versions
		m.address = common.HexToAddress(string(bytes))
	} else {
		m.address = common.HexToAddress(af.Address)
		m.observe = af.Observe
	}
	return m.address, nil
}

// Get the cached address
func (m *AddressManager) GetAddress() common.Address {
	return m.address
}

// Get the cached observe flag
func (m *AddressManager) IsObserve() bool {
	return m.observe
}

// Sets the node address and observe flag, and saves both to disk
func (m *AddressManager) SetAndSaveAddress(newAddress common.Address, observe bool) error {
	m.address = newAddress
	m.observe = observe
	af := addressFile{Address: newAddress.Hex(), Observe: observe}
	bytes, err := json.Marshal(af)
	if err != nil {
		return fmt.Errorf("error encoding address file: %w", err)
	}
	if err := os.WriteFile(m.path, bytes, addressFileMode); err != nil {
		return fmt.Errorf("error writing address file [%s] to disk: %w", m.path, err)
	}
	return nil
}

// Delete the saved address file from disk
func (m *AddressManager) DeleteAddressFile() error {
	err := os.Remove(m.path)
	if err != nil {
		return fmt.Errorf("error deleting address file [%s]: %w", m.path, err)
	}
	m.address = common.Address{}
	m.observe = false
	return nil
}

// CheckObserveMode reads the address file at path and returns true if observe mode is active.
// Returns false if the file doesn't exist, can't be read, or observe is not set.
func CheckObserveMode(path string) bool {
	am := NewAddressManager(path)
	if _, err := am.LoadAddress(); err != nil {
		return false
	}
	return am.IsObserve()
}
