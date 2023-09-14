package wallet

import (
	"context"
	"fmt"
	"math/big"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
)

// Get a transactor for the node account
func (w *LocalWallet) GetNodeAccountTransactor() (*bind.TransactOpts, error) {
	status := w.GetStatus()
	switch status {
	case WalletStatus_NoAddress:
		return nil, fmt.Errorf("node wallet does not have an address loaded - please create or recover a node wallet")
	case WalletStatus_NoKeystore:
		return nil, fmt.Errorf("node wallet is in read-only mode; it cannot transact because no keystore is loaded")
	case WalletStatus_NoPassword:
		return nil, fmt.Errorf("node wallet is in read-only mode; no password is loaded for the wallet")
	case WalletStatus_KeystoreMismatch:
		return nil, fmt.Errorf("node wallet is in read-only mode; the keystore is for a different wallet than the one it is using")
	case WalletStatus_Ready:
		transactor, err := bind.NewKeyedTransactorWithChainID(w.nodePrivateKey, w.chainID)
		transactor.Context = context.Background()
		return transactor, err
	default:
		return nil, fmt.Errorf("unknown wallet status %v", status)
	}
}

// Get the node account private key bytes
func (w *LocalWallet) GetNodePrivateKeyBytes() []byte {
	// Return private key bytes
	return crypto.FromECDSA(w.nodePrivateKey)
}

// Get the derived key & derivation path for the node account at the index
func (w *LocalWallet) getNodeDerivedKey(index uint) (*hdkeychain.ExtendedKey, string, error) {
	// Get the derivation path
	if w.keystoreManager.data.DerivationPath == "" {
		w.keystoreManager.data.DerivationPath = DefaultNodeKeyPath
	}
	derivationPath := fmt.Sprintf(w.keystoreManager.data.DerivationPath, index)

	// Parse derivation path
	path, err := accounts.ParseDerivationPath(derivationPath)
	if err != nil {
		return nil, "", fmt.Errorf("Invalid node key derivation path '%s': %w", derivationPath, err)
	}

	// Follow derivation path
	key := w.masterKey
	for i, n := range path {
		// Use the legacy implementation for Goerli
		// TODO: remove this if Prater ever goes away!
		if w.chainID.Cmp(big.NewInt(5)) == 0 {
			key, err = key.DeriveNonStandard(n)
		} else {
			key, err = key.Derive(n)
		}
		if err == hdkeychain.ErrInvalidChild {
			return w.getNodeDerivedKey(index + 1)
		} else if err != nil {
			return nil, "", fmt.Errorf("Invalid child key at depth %d: %w", i, err)
		}
	}

	// Return
	return key, derivationPath, nil
}
