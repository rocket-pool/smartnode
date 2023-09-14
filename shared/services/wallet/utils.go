package wallet

import (
	"fmt"
)

// Recover a wallet keystore from a mnemonic - only used for testing mnemonics
func TestRecovery(derivationPath string, walletIndex uint, mnemonic string, chainID uint) (*LocalWallet, error) {
	// Create a new dummy wallet
	w, err := NewLocalWallet("", "", "", chainID, false)
	if err != nil {
		return nil, fmt.Errorf("error creating new test node wallet: %w", err)
	}
	w.passwordManager.SetAndSave([]byte("test password"))

	err = w.Recover(derivationPath, walletIndex, mnemonic)
	if err != nil {
		return nil, fmt.Errorf("error test recovering mnemonic: %w", err)
	}
	return w, nil
}
