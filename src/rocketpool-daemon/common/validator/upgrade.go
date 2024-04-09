package validator

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	nodewallet "github.com/rocket-pool/node-manager-core/node/wallet"
	"github.com/rocket-pool/node-manager-core/wallet"
)

// Keystore for node wallets
type legacyWallet struct {
	Crypto         map[string]interface{} `json:"crypto"`
	Name           string                 `json:"name"`
	Version        uint                   `json:"version"`
	UUID           uuid.UUID              `json:"uuid"`
	DerivationPath string                 `json:"derivationPath,omitempty"`
	WalletIndex    uint                   `json:"walletIndex,omitempty"`
	NextAccount    uint                   `json:"next_account"`
}

// Check for a legacy wallet and upgrade it if necessary, returning true if an upgrade was performed
func CheckAndUpgradeWallet(walletDataPath string, nextAccountPath string, logger *slog.Logger) (bool, error) {
	// Check if there is a wallet file
	_, err := os.Stat(walletDataPath)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("error checking for presence of wallet file [%s]: %w", walletDataPath, err)
	}

	// Read the file
	bytes, err := os.ReadFile(walletDataPath)
	if err != nil {
		return false, fmt.Errorf("error reading wallet data at [%s]: %w", walletDataPath, err)
	}

	// Check if it's a legacy wallet
	legacyWallet := new(legacyWallet)
	err = json.Unmarshal(bytes, legacyWallet)
	if err != nil || len(legacyWallet.Crypto) == 0 {
		// Nothing to decode so this isn't a legacy wallet
		return false, nil
	}

	// Convert to the new form
	logger.Info("Legacy wallet detected, upgrading...")
	logger.Debug("Legacy wallet info", slog.Uint64("next", uint64(legacyWallet.NextAccount)), slog.String("path", legacyWallet.DerivationPath), slog.Uint64("index", uint64(legacyWallet.WalletIndex)))
	newWallet := wallet.WalletData{
		Type: wallet.WalletType_Local,
		LocalData: wallet.LocalWalletData{
			Crypto:         legacyWallet.Crypto,
			Name:           legacyWallet.Name,
			Version:        legacyWallet.Version,
			UUID:           legacyWallet.UUID,
			DerivationPath: legacyWallet.DerivationPath,
			WalletIndex:    legacyWallet.WalletIndex,
		},
		HardwareData: wallet.HardwareWalletData{},
	}

	// Save the next account data
	err = saveNextAccount(uint64(legacyWallet.NextAccount), nextAccountPath)
	if err != nil {
		return false, err
	}

	// Serialize the new wallet
	bytes, err = json.Marshal(newWallet)
	if err != nil {
		return false, fmt.Errorf("error serializing new wallet data: %w", err)
	}

	// Write the file
	err = os.WriteFile(walletDataPath, bytes, nodewallet.FileMode)
	if err != nil {
		return false, fmt.Errorf("error writing wallet data to [%s]: %w", walletDataPath, err)
	}

	// Done
	logger.Info("Wallet upgrade complete.")
	return true, nil
}
