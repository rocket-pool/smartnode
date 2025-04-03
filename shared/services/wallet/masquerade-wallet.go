package wallet

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"os"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/rocket-pool/smartnode/shared/services/passwords"
	"github.com/rocket-pool/smartnode/shared/services/wallet/keystore"
	"github.com/tyler-smith/go-bip39"
	eth2ks "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
)

var ErrIsMasquerading = errors.New("The node is currently masquerading. Restore the node wallet using 'rocketpool wallet restore-address'")

// masqueradeWallet
type masqueradeWallet struct {

	// Core
	walletPath string
	pm         *passwords.PasswordManager
	am         *AddressManager
	encryptor  *eth2ks.Encryptor
	chainID    *big.Int

	// Encrypted store
	ws *walletStore

	// Seed & master key
	seed []byte
	mk   *hdkeychain.ExtendedKey

	// Node key cache
	nodeKey     *ecdsa.PrivateKey
	nodeKeyPath string
}

// Gets the derived wallet address, if one is loaded
func (w *masqueradeWallet) GetAddress() (common.Address, error) {

	// Check wallet is initialized
	if !w.IsInitialized() {
		return common.Address{}, errors.New("wallet is not initialized")
	}

	// Get private key
	privateKey, _, err := w.getNodePrivateKey()
	if err != nil {
		return common.Address{}, errors.New("Could not get node private key")
	}

	// Get public key
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return common.Address{}, errors.New("Could not get node public key)")
	}
	return crypto.PubkeyToAddress(*publicKeyECDSA), nil
}

// Masquerade as another node address, running all node functions as that address (in read only mode)
func (w *masqueradeWallet) MasqueradeAsAddress(newAddress common.Address) error {
	return w.masqueradeImpl(newAddress)
}

// End masquerading as another node address, and use the wallet's address (returning to read/write mode)
func (w *masqueradeWallet) RestoreAddressToWallet() error {
	if w.am == nil {
		return errors.New("wallet is not loaded")
	}

	return w.am.DeleteAddressFile()
}

// Gets the wallet's chain ID
func (w *masqueradeWallet) GetChainID() *big.Int {
	copy := big.NewInt(0).Set(w.chainID)
	return copy
}

// Add a keystore to the wallet
func (w *masqueradeWallet) AddKeystore(name string, ks keystore.Keystore) {
	return
}

// Check if the wallet has been initialized
func (w *masqueradeWallet) IsInitialized() bool {
	return (w.ws != nil && w.seed != nil && w.mk != nil)
}

// Attempt to initialize the wallet if not initialized and return status
func (w *masqueradeWallet) GetInitialized() (bool, error) {
	if w.IsInitialized() {
		return true, nil
	}
	return w.loadStore()
}

// Serialize the wallet to a JSON string
func (w *masqueradeWallet) String() (string, error) {
	return "", ErrIsMasquerading

}

// Initialize the wallet from a random seed
func (w *masqueradeWallet) Initialize(derivationPath string, walletIndex uint) (string, error) {
	return "", ErrIsMasquerading

}

// Recover a wallet from a mnemonic
func (w *masqueradeWallet) Recover(derivationPath string, walletIndex uint, mnemonic string) error {
	return ErrIsMasquerading

}

// Recover a wallet from a mnemonic - only used for testing mnemonics
func (w *masqueradeWallet) TestRecovery(derivationPath string, walletIndex uint, mnemonic string) error {

	// Check mnemonic
	if !bip39.IsMnemonicValid(mnemonic) {
		return fmt.Errorf("Invalid mnemonic '%s'", mnemonic)
	}

	// Generate seed
	w.seed = bip39.NewSeed(mnemonic, "")

	// Create master key
	var err error
	w.mk, err = hdkeychain.NewMaster(w.seed, &chaincfg.MainNetParams)
	if err != nil {
		return fmt.Errorf("Could not create wallet master key: %w", err)
	}

	// Create wallet store
	w.ws = &walletStore{
		Name:           w.encryptor.Name(),
		Version:        w.encryptor.Version(),
		UUID:           uuid.New(),
		DerivationPath: derivationPath,
		WalletIndex:    walletIndex,
		NextAccount:    0,
	}

	// Return
	return nil

}

// Save the wallet store to disk
func (w *masqueradeWallet) Save() error {
	return ErrIsMasquerading
}

// Delete the wallet store from disk
func (w *masqueradeWallet) Delete() error {
	return ErrIsMasquerading
}

// Signs a serialized TX using the wallet's private key
func (w *masqueradeWallet) Sign(serializedTx []byte) ([]byte, error) {
	return nil, ErrIsMasquerading
}

// Signs an arbitrary message using the wallet's private key
func (w *masqueradeWallet) SignMessage(message string) ([]byte, error) {
	return nil, ErrIsMasquerading
}

// Reloads wallet from disk
func (w *masqueradeWallet) Reload() error {
	_, err := w.loadStore()
	return err
}

// Load the wallet store from disk and decrypt it
func (w *masqueradeWallet) loadStore() (bool, error) {

	// Read wallet store from disk; cancel if not found
	wsBytes, err := os.ReadFile(w.walletPath)
	if err != nil {
		return false, nil
	}

	// Decode wallet store
	w.ws = new(walletStore)
	if err = json.Unmarshal(wsBytes, w.ws); err != nil {
		return false, fmt.Errorf("Could not decode wallet: %w", err)
	}

	// Upgrade legacy wallets to include derivation paths
	if w.ws.DerivationPath == "" {
		w.ws.DerivationPath = DefaultNodeKeyPath
	}

	// Get wallet password
	password, err := w.pm.GetPassword()
	if err != nil {
		return false, fmt.Errorf("Could not get wallet password: %w", err)
	}

	// Decrypt seed
	w.seed, err = w.encryptor.Decrypt(w.ws.Crypto, password)
	if err != nil {
		return false, fmt.Errorf("Could not decrypt wallet seed: %w", err)
	}

	// Create master key
	w.mk, err = hdkeychain.NewMaster(w.seed, &chaincfg.MainNetParams)
	if err != nil {
		return false, fmt.Errorf("Could not create wallet master key: %w", err)
	}

	// Load the node address
	_, _, err = w.am.LoadAddress()
	if err != nil {
		return false, fmt.Errorf("Could not load node address: %w", err)
	}

	// Return
	return true, nil

}

// Initialize the encrypted wallet store from a mnemonic
func (w *masqueradeWallet) initializeStore(_ string, _ uint, _ string) error {
	return ErrIsMasquerading
}

// Masquerade as another node address, running all node functions as that address (in read only mode)
func (w *masqueradeWallet) masqueradeImpl(newAddress common.Address) error {
	return w.am.SetAndSaveAddress(newAddress)
}
