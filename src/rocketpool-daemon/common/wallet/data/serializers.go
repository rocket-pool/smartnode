package data

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/wallet/keystore"
)

// Interface for serializing and deserializing arbitrary data to disk
type dataSerializer[dataType any] interface {
	serialize(dataType) ([]byte, error)
	deserialize([]byte) (dataType, error)
}

/// ======================
/// === Wallet Address ===
/// ======================

// File signature for storing the node address on disk
type walletAddressFile struct {
	Address common.Address `json:"address"`
}

type walletAddressSerializer struct {
}

func (s walletAddressSerializer) serialize(address common.Address) ([]byte, error) {
	// Store the address in a JSON wrapper
	file := walletAddressFile{
		Address: address,
	}
	return json.Marshal(file)
}

func (s walletAddressSerializer) deserialize(bytes []byte) (common.Address, error) {
	// Load the address from the JSON wrapper
	file := new(walletAddressFile)
	err := json.Unmarshal(bytes, file)
	if err != nil {
		return common.Address{}, err
	}
	return file.Address, nil
}

/// =======================
/// === Wallet Keystore ===
/// =======================

type walletKeystoreSerializer struct {
}

func (s walletKeystoreSerializer) serialize(file *keystore.WalletKeystore) ([]byte, error) {
	return json.Marshal(file)
}

func (s walletKeystoreSerializer) deserialize(bytes []byte) (*keystore.WalletKeystore, error) {
	file := new(keystore.WalletKeystore)
	err := json.Unmarshal(bytes, file)
	if err != nil {
		return nil, err
	}
	return file, nil
}

/// ================
/// === Password ===
/// ================

type passwordSerializer struct {
}

func (s passwordSerializer) serialize(password []byte) ([]byte, error) {
	// No serialization required, the password file stores the raw ASCII value
	return password, nil
}

func (s passwordSerializer) deserialize(bytes []byte) ([]byte, error) {
	// No deserialization required, the password file stores the raw ASCII value
	return bytes, nil
}
