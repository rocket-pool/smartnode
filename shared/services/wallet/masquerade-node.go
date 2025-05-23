package wallet

import (
	"context"
	"crypto/ecdsa"
	"fmt"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

// Get the node account
func (w *masqueradeWallet) GetNodeAccount() (accounts.Account, error) {
	address, err := w.am.LoadAddress()
	if err != nil {
		return accounts.Account{}, fmt.Errorf("Could not load node address: %w", err)
	}

	// Create & return account
	return accounts.Account{
		Address: address,
		URL: accounts.URL{
			Scheme: "",
			Path:   "",
		},
	}, nil

}

// Get a transactor for the masqueraded node account. There is no private key so transactions will fail
func (w *masqueradeWallet) GetNodeAccountTransactor() (*bind.TransactOpts, error) {
	// Masqueraded account
	account, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Create & return transactor
	transactor := &bind.TransactOpts{}
	transactor.GasFeeCap = w.maxFee
	transactor.GasTipCap = w.maxPriorityFee
	transactor.GasLimit = w.gasLimit
	transactor.Context = context.Background()
	transactor.NoSend = true
	transactor.From = account.Address
	return transactor, err
}

// Get the node account private key bytes
func (w *masqueradeWallet) GetNodePrivateKeyBytes() ([]byte, error) {
	return nil, ErrIsMasquerading

}

// Get the node private key
func (w *masqueradeWallet) getNodePrivateKey() (*ecdsa.PrivateKey, string, error) {

	// Check for cached node key
	if w.nodeKey != nil {
		return w.nodeKey, w.nodeKeyPath, nil
	}

	// Get derived key
	derivedKey, path, err := w.getNodeDerivedKey(w.ws.WalletIndex)
	if err != nil {
		return nil, "", err
	}

	// Get private key
	privateKey, err := derivedKey.ECPrivKey()
	if err != nil {
		return nil, "", fmt.Errorf("Could not get node private key: %w", err)
	}
	privateKeyECDSA := privateKey.ToECDSA()

	// Cache node key
	w.nodeKey = privateKeyECDSA
	w.nodeKeyPath = path

	// Return
	return privateKeyECDSA, path, nil

}

// Get the derived key & derivation path for the node account at the index
func (w *masqueradeWallet) getNodeDerivedKey(index uint) (*hdkeychain.ExtendedKey, string, error) {

	// Get derivation path
	if w.ws.DerivationPath == "" {
		w.ws.DerivationPath = DefaultNodeKeyPath
	}
	derivationPath := fmt.Sprintf(w.ws.DerivationPath, index)

	// Parse derivation path
	path, err := accounts.ParseDerivationPath(derivationPath)
	if err != nil {
		return nil, "", fmt.Errorf("Invalid node key derivation path '%s': %w", derivationPath, err)
	}

	// Follow derivation path
	key := w.mk
	for i, n := range path {
		key, err = key.Derive(n)
		if err == hdkeychain.ErrInvalidChild {
			return w.getNodeDerivedKey(index + 1)
		} else if err != nil {
			return nil, "", fmt.Errorf("Invalid child key at depth %d: %w", i, err)
		}
	}

	// Return
	return key, derivationPath, nil

}
