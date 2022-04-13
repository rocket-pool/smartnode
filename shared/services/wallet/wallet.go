package wallet

import (
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/google/uuid"
	"github.com/tyler-smith/go-bip39"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
	eth2ks "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"

	"github.com/rocket-pool/smartnode/shared/services/passwords"
	"github.com/rocket-pool/smartnode/shared/services/wallet/fees"
	"github.com/rocket-pool/smartnode/shared/services/wallet/keystore"
)

// Config
const (
	EntropyBits              = 256
	FileMode                 = 0600
	DefaultNodeKeyPath       = "m/44'/60'/0'/0/%d"
	LedgerLiveNodeKeyPath    = "m/44'/60'/%d/0/0"
	MyEtherWalletNodeKeyPath = "m/44'/60'/0'/%d"
)

// Wallet
type Wallet struct {

	// Core
	walletPath string
	pm         *passwords.PasswordManager
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

	// Validator key caches
	validatorKeys       map[uint]*eth2types.BLSPrivateKey
	validatorKeyIndices map[string]uint

	// Keystores
	keystores map[string]keystore.Keystore

	// Fee recipient managers
	feeRecipientManagers map[string]fees.FeeRecipientManager

	// Desired gas price & limit from config
	maxFee         *big.Int
	maxPriorityFee *big.Int
	gasLimit       uint64
}

// Encrypted wallet store
type walletStore struct {
	Crypto         map[string]interface{} `json:"crypto"`
	Name           string                 `json:"name"`
	Version        uint                   `json:"version"`
	UUID           uuid.UUID              `json:"uuid"`
	DerivationPath string                 `json:"derivationPath,omitempty"`
	NextAccount    uint                   `json:"next_account"`
}

// Create new wallet
func NewWallet(walletPath string, chainId uint, maxFee *big.Int, maxPriorityFee *big.Int, gasLimit uint64, passwordManager *passwords.PasswordManager) (*Wallet, error) {

	// Initialize wallet
	w := &Wallet{
		walletPath:           walletPath,
		pm:                   passwordManager,
		encryptor:            eth2ks.New(),
		chainID:              big.NewInt(int64(chainId)),
		validatorKeys:        map[uint]*eth2types.BLSPrivateKey{},
		validatorKeyIndices:  map[string]uint{},
		keystores:            map[string]keystore.Keystore{},
		feeRecipientManagers: map[string]fees.FeeRecipientManager{},
		maxFee:               maxFee,
		maxPriorityFee:       maxPriorityFee,
		gasLimit:             gasLimit,
	}

	// Load & decrypt wallet store
	if _, err := w.loadStore(); err != nil {
		return nil, err
	}

	// Return
	return w, nil

}

// Gets the wallet's chain ID
func (w *Wallet) GetChainID() *big.Int {
	copy := big.NewInt(0).Set(w.chainID)
	return copy
}

// Add a keystore to the wallet
func (w *Wallet) AddKeystore(name string, ks keystore.Keystore) {
	w.keystores[name] = ks
}

// Add a fee recipient manager to the wallet
func (w *Wallet) AddFeeRecipientManager(name string, fm fees.FeeRecipientManager) {
	w.feeRecipientManagers[name] = fm
}

// Check if the wallet has been initialized
func (w *Wallet) IsInitialized() bool {
	return (w.ws != nil && w.seed != nil && w.mk != nil)
}

// Attempt to initialize the wallet if not initialized and return status
func (w *Wallet) GetInitialized() (bool, error) {
	if w.IsInitialized() {
		return true, nil
	}
	return w.loadStore()
}

// Serialize the wallet to a JSON string
func (w *Wallet) String() (string, error) {

	// Check wallet is initialized
	if !w.IsInitialized() {
		return "", errors.New("Wallet is not initialized")
	}

	// Encode wallet store
	wsBytes, err := json.Marshal(w.ws)
	if err != nil {
		return "", fmt.Errorf("Could not encode wallet: %w", err)
	}

	// Return
	return string(wsBytes), nil

}

// Initialize the wallet from a random seed
func (w *Wallet) Initialize(derivationPath string) (string, error) {

	// Check wallet is not initialized
	if w.IsInitialized() {
		return "", errors.New("Wallet is already initialized")
	}

	// Generate mnemonic entropy
	entropy, err := bip39.NewEntropy(EntropyBits)
	if err != nil {
		return "", fmt.Errorf("Could not generate wallet mnemonic entropy bytes: %w", err)
	}

	// Generate mnemonic
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", fmt.Errorf("Could not generate wallet mnemonic: %w", err)
	}

	// Initialize wallet store
	if err := w.initializeStore(derivationPath, mnemonic); err != nil {
		return "", err
	}

	// Return
	return mnemonic, nil

}

// Recover a wallet from a mnemonic
func (w *Wallet) Recover(derivationPath string, mnemonic string) error {

	// Check wallet is not initialized
	if w.IsInitialized() {
		return errors.New("Wallet is already initialized")
	}

	// Check mnemonic
	if !bip39.IsMnemonicValid(mnemonic) {
		return fmt.Errorf("Invalid mnemonic '%s'", mnemonic)
	}

	// Initialize wallet store
	if err := w.initializeStore(derivationPath, mnemonic); err != nil {
		return err
	}

	// Return
	return nil

}

// Save the wallet store to disk
func (w *Wallet) Save() error {

	// Check wallet is initialized
	if !w.IsInitialized() {
		return errors.New("Wallet is not initialized")
	}

	// Encode wallet store
	wsBytes, err := json.Marshal(w.ws)
	if err != nil {
		return fmt.Errorf("Could not encode wallet: %w", err)
	}

	// Write wallet store to disk
	if err := ioutil.WriteFile(w.walletPath, wsBytes, FileMode); err != nil {
		return fmt.Errorf("Could not write wallet to disk: %w", err)
	}

	// Return
	return nil

}

// Reloads wallet from disk
func (w *Wallet) Reload() error {
	_, err := w.loadStore()
	return err
}

// Load the wallet store from disk and decrypt it
func (w *Wallet) loadStore() (bool, error) {

	// Read wallet store from disk; cancel if not found
	wsBytes, err := ioutil.ReadFile(w.walletPath)
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

	// Return
	return true, nil

}

// Initialize the encrypted wallet store from a mnemonic
func (w *Wallet) initializeStore(derivationPath string, mnemonic string) error {

	// Generate seed
	w.seed = bip39.NewSeed(mnemonic, "")

	// Create master key
	var err error
	w.mk, err = hdkeychain.NewMaster(w.seed, &chaincfg.MainNetParams)
	if err != nil {
		return fmt.Errorf("Could not create wallet master key: %w", err)
	}

	// Get wallet password
	password, err := w.pm.GetPassword()
	if err != nil {
		return fmt.Errorf("Could not get wallet password: %w", err)
	}

	// Encrypt seed
	encryptedSeed, err := w.encryptor.Encrypt(w.seed, password)
	if err != nil {
		return fmt.Errorf("Could not encrypt wallet seed: %w", err)
	}

	// Create wallet store
	w.ws = &walletStore{
		Crypto:         encryptedSeed,
		Name:           w.encryptor.Name(),
		Version:        w.encryptor.Version(),
		UUID:           uuid.New(),
		DerivationPath: derivationPath,
		NextAccount:    0,
	}

	// Return
	return nil

}
