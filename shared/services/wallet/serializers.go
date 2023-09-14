package wallet

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
)

// Interface for serializing and deserializing arbitrary data to disk
type FileSerializer[DataType any] interface {
	Serialize(DataType) ([]byte, error)
	Deserialize([]byte) (DataType, error)
}

/// ======================
/// === Wallet Address ===
/// ======================

// File signature for storing the node address on disk
type WalletAddressFile struct {
	Address common.Address `json:"address"`
}

type WalletAddressSerializer struct {
}

func (s WalletAddressSerializer) Serialize(address common.Address) ([]byte, error) {
	// Store the address in a JSON wrapper
	file := WalletAddressFile{
		Address: address,
	}
	return json.Marshal(file)
}

func (s WalletAddressSerializer) Deserialize(bytes []byte) (common.Address, error) {
	// Load the address from the JSON wrapper
	file := new(WalletAddressFile)
	err := json.Unmarshal(bytes, file)
	if err != nil {
		return common.Address{}, err
	}
	return file.Address, nil
}

/// =======================
/// === Wallet Keystore ===
/// =======================

type WalletKeystoreFile struct {
	Crypto         map[string]interface{} `json:"crypto"`
	Name           string                 `json:"name"`
	Version        uint                   `json:"version"`
	UUID           uuid.UUID              `json:"uuid"`
	DerivationPath string                 `json:"derivationPath,omitempty"`
	WalletIndex    uint                   `json:"walletIndex,omitempty"`
	NextAccount    uint                   `json:"next_account"`
}

type WalletKeystoreSerializer struct {
}

func (s WalletKeystoreSerializer) Serialize(file *WalletKeystoreFile) ([]byte, error) {
	return json.Marshal(file)
}

func (s WalletKeystoreSerializer) Deserialize(bytes []byte) (*WalletKeystoreFile, error) {
	file := new(WalletKeystoreFile)
	err := json.Unmarshal(bytes, file)
	if err != nil {
		return nil, err
	}
	return file, nil
}

/// ================
/// === Password ===
/// ================

type PasswordSerializer struct {
}

func (s PasswordSerializer) Serialize(password []byte) ([]byte, error) {
	// No serialization required, the password file stores the raw ASCII value
	return password, nil
}

func (s PasswordSerializer) Deserialize(bytes []byte) ([]byte, error) {
	// No deserialization required, the password file stores the raw ASCII value
	return bytes, nil
}
