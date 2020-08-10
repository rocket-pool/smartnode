package wallet

import (
    "crypto/ecdsa"
    "errors"
    "fmt"

    "github.com/ethereum/go-ethereum/accounts"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/crypto"
)


// Config
const NodeKeyPath = "m/44'/60'/0'/0/0"


// Get the node account
func (w *Wallet) GetNodeAccount() (accounts.Account, error) {

    // Check wallet is initialized
    if !w.IsInitialized() {
        return accounts.Account{}, errors.New("Wallet is not initialized")
    }

    // Get private key
    privateKey, err := w.deriveNodeKey()
    if err != nil {
        return accounts.Account{}, err
    }

    // Get public key
    publicKey := privateKey.Public()
    publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
    if !ok {
        return accounts.Account{}, errors.New("Could not get node public key")
    }

    // Create & return account
    return accounts.Account{
        Address: crypto.PubkeyToAddress(*publicKeyECDSA),
        URL: accounts.URL{
            Scheme: "",
            Path: NodeKeyPath,
        },
    }, nil

}


// Get a transactor for the node account
func (w *Wallet) GetNodeAccountTransactor() (*bind.TransactOpts, error) {

    // Check wallet is initialized
    if !w.IsInitialized() {
        return nil, errors.New("Wallet is not initialized")
    }

    // Get private key
    privateKey, err := w.deriveNodeKey()
    if err != nil {
        return nil, err
    }

    // Create & return transactor
    return bind.NewKeyedTransactor(privateKey), nil

}


// Derive the node private key
func (w *Wallet) deriveNodeKey() (*ecdsa.PrivateKey, error) {

    // Parse node key derivation path
    path, err := accounts.ParseDerivationPath(NodeKeyPath)
    if err != nil {
        return nil, fmt.Errorf("Invalid node account derivation path: %w", err)
    }

    // Follow derivation path
    key := w.mk
    for i, n := range path {
        key, err = key.Child(n)
        if err != nil {
            return nil, fmt.Errorf("Invalid child key at depth %d: %w", i, err)
        }
    }

    // Get private key
    privateKey, err := key.ECPrivKey()
    if err != nil {
        return nil, fmt.Errorf("Could not get node private key: %w", err)
    }

    // Return
    return privateKey.ToECDSA(), nil

}

