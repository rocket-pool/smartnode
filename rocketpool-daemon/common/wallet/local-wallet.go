package wallet

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/tyler-smith/go-bip39"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
	eth2ks "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"

	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/wallet/data"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/wallet/keystore"
	sharedtypes "github.com/rocket-pool/smartnode/shared/types"
)

// Config
const (
	EntropyBits              = 256
	FileMode                 = 0600
	DefaultNodeKeyPath       = "m/44'/60'/0'/0/%d"
	LedgerLiveNodeKeyPath    = "m/44'/60'/%d/0/0"
	MyEtherWalletNodeKeyPath = "m/44'/60'/0'/%d"
)

// LocalWallet
type LocalWallet struct {
	// Managers
	addressManager  *data.DataManager[common.Address]
	keystoreManager *data.DataManager[*keystore.WalletKeystore]
	passwordManager *data.DataManager[[]byte]

	// Node private key info
	encryptor      *eth2ks.Encryptor
	seed           []byte
	masterKey      *hdkeychain.ExtendedKey
	nodePrivateKey *ecdsa.PrivateKey
	nodeKeyPath    string

	// Validator keys
	validatorKeys      map[uint]*eth2types.BLSPrivateKey
	validatorKeystores map[string]keystore.Keystore

	// Misc cache
	chainID *big.Int
}

// Create new wallet
func NewLocalWallet(walletKeystorePath string, walletAddressPath string, passwordFilePath string, chainID uint, init bool) (*LocalWallet, error) {
	// Create the wallet
	w := &LocalWallet{
		// Create managers
		addressManager:  data.NewAddressManager(walletAddressPath),
		keystoreManager: data.NewKeystoreManager(walletKeystorePath),
		passwordManager: data.NewPasswordManager(passwordFilePath),

		// Initialize other fields
		encryptor:          eth2ks.New(),
		validatorKeys:      map[uint]*eth2types.BLSPrivateKey{},
		validatorKeystores: map[string]keystore.Keystore{},
		chainID:            big.NewInt(int64(chainID)),
	}

	// Initialize it
	if init {
		// Load the files from disk
		_, addressFileExists, err := w.addressManager.InitializeData()
		if err != nil {
			return nil, fmt.Errorf("error getting wallet address: %w", err)
		}
		keystore, keystoreFileExists, err := w.keystoreManager.InitializeData()
		if err != nil {
			return nil, fmt.Errorf("error getting wallet keystore: %w", err)
		}
		password, passwordFileExists, err := w.passwordManager.InitializeData()
		if err != nil {
			return nil, fmt.Errorf("error getting wallet password: %w", err)
		}

		// Load the keystore if possible and compare it to the node address
		if keystoreFileExists {
			// If there's no password, don't load the keystore
			if !passwordFileExists {
				return w, nil
			}

			// Load the keystore, saving the address file if it doesn't exist
			err = w.loadKeyFromKeystore(keystore, password, !addressFileExists)
			if err != nil {
				return nil, fmt.Errorf("error loading wallet key from keystore: %w", err)
			}

		}
		if err != nil {
			return nil, fmt.Errorf("error initializing wallet: %w", err)
		}
	}
	return w, nil
}

// Gets the wallet's chain ID
func (w *LocalWallet) GetChainID() *big.Int {
	copy := big.NewInt(0).Set(w.chainID)
	return copy
}

// Gets the status of the wallet and its artifacts
func (w *LocalWallet) GetStatus() sharedtypes.WalletStatus {
	// Get the data and its existence
	status := sharedtypes.WalletStatus{}
	status.NodeAddress, status.HasAddress = w.addressManager.Get()
	_, status.HasKeystore = w.keystoreManager.Get()
	_, status.HasPassword = w.passwordManager.Get()

	if status.HasKeystore && status.HasPassword {
		status.KeystoreAddress = crypto.PubkeyToAddress(w.nodePrivateKey.PublicKey)
	}
	status.IsPasswordSaved = w.passwordManager.IsSavedToDisk()
	return status
}

// Get the wallet's address, if one is loaded
func (w *LocalWallet) GetAddress() (common.Address, bool) {
	return w.addressManager.Get()
}

// Add a validator keystore to the wallet
func (w *LocalWallet) AddValidatorKeystore(name string, ks keystore.Keystore) {
	w.validatorKeystores[name] = ks
}

// Serialize the wallet keystore to a JSON string
func (w *LocalWallet) String() (string, error) {
	// Encode the wallet keystore
	keystoreString, isSet, err := w.keystoreManager.String()
	if err != nil {
		return "", fmt.Errorf("error serializing wallet keystore into a string: %w", err)
	}
	if !isSet {
		return "", fmt.Errorf("wallet keystore has not been set yet")
	}

	// Return
	return keystoreString, nil
}

// Initialize the wallet from a random seed
func (w *LocalWallet) CreateNewWallet(derivationPath string, walletIndex uint) (string, error) {
	if w.keystoreManager.HasValue() {
		return "", fmt.Errorf("wallet keystore is already present - please delete it before creating a new wallet")
	}

	// Generate random entropy for the mnemonic
	entropy, err := bip39.NewEntropy(EntropyBits)
	if err != nil {
		return "", fmt.Errorf("error generating wallet mnemonic entropy bytes: %w", err)
	}

	// Generate a new mnemonic
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", fmt.Errorf("error generating wallet mnemonic: %w", err)
	}

	// Initialize the wallet with it
	err = w.initializeKeystore(derivationPath, walletIndex, mnemonic)
	if err != nil {
		return "", fmt.Errorf("error initializing new wallet keystore: %w", err)
	}
	return mnemonic, nil
}

// Recover a wallet from a mnemonic
func (w *LocalWallet) Recover(derivationPath string, walletIndex uint, mnemonic string) error {
	if w.keystoreManager.HasValue() {
		return fmt.Errorf("wallet keystore is already present - please delete it before recovering an existing wallet")
	}

	// Check the mnemonic
	if !bip39.IsMnemonicValid(mnemonic) {
		return fmt.Errorf("invalid mnemonic '%s'", mnemonic)
	}

	// Initialize the wallet with it
	err := w.initializeKeystore(derivationPath, walletIndex, mnemonic)
	if err != nil {
		return fmt.Errorf("error initializing wallet keystore with recovered data: %w", err)
	}
	return nil
}

// Stores a new password in memory but does not save it to disk, then reloads the keystore and corresponding details
func (w *LocalWallet) RememberPassword(password []byte) {
	w.passwordManager.Set(password)
}

// Removes the wallet's password from memory and invalidates the keystore so it can no longer transact
func (w *LocalWallet) ForgetPassword() {
	w.passwordManager.Clear()
}

// Retrieves the wallet's password
func (w *LocalWallet) GetPassword() ([]byte, bool) {
	return w.passwordManager.Get()
}

// Save the wallet's password to disk
func (w *LocalWallet) SavePassword() error {
	err := w.passwordManager.Save()
	if err != nil {
		return fmt.Errorf("error saving wallet password: %w", err)
	}
	return nil
}

// Delete the wallet password from disk, but retain it in memory
func (w *LocalWallet) DeletePassword() error {
	err := w.passwordManager.Delete()
	if err != nil {
		return fmt.Errorf("error deleting wallet password: %w", err)
	}
	return nil
}

// Delete the wallet keystore from disk and purge it from memory
func (w *LocalWallet) DeleteKeystore() error {
	w.keystoreManager.Clear()
	w.seed = []byte{}
	w.masterKey = nil
	w.nodePrivateKey = nil
	w.nodeKeyPath = ""
	err := w.keystoreManager.Delete()
	if err != nil {
		return fmt.Errorf("error deleting wallet keystore: %w", err)
	}
	return nil
}

// Get a transactor for the wallet
func (w *LocalWallet) GetTransactor() (*bind.TransactOpts, error) {
	status := w.GetStatus()
	if !status.HasAddress {
		return nil, fmt.Errorf("node wallet does not have an address loaded - please create or recover a node wallet")
	}
	if !status.HasKeystore {
		return nil, fmt.Errorf("node wallet is in read-only mode; it cannot transact because no keystore is loaded")
	}
	if !status.HasPassword {
		return nil, fmt.Errorf("node wallet is in read-only mode; no password is loaded for the wallet")
	}
	if status.NodeAddress != status.KeystoreAddress {
		return nil, fmt.Errorf("node wallet is in read-only mode; the keystore is for a different wallet than the one it is using")
	}

	transactor, err := bind.NewKeyedTransactorWithChainID(w.nodePrivateKey, w.chainID)
	transactor.Context = context.Background()
	return transactor, err
}

// Get the node account private key bytes
func (w *LocalWallet) GetNodePrivateKeyBytes() []byte {
	// Return private key bytes
	return crypto.FromECDSA(w.nodePrivateKey)
}

// Signs a serialized TX using the wallet's private key
func (w *LocalWallet) Sign(serializedTx []byte) ([]byte, error) {
	tx := types.Transaction{}
	err := tx.UnmarshalBinary(serializedTx)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling TX: %w", err)
	}

	signer := types.NewLondonSigner(w.chainID)
	signedTx, err := types.SignTx(&tx, signer, w.nodePrivateKey)
	if err != nil {
		return nil, fmt.Errorf("Error signing TX: %w", err)
	}

	signedData, err := signedTx.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("Error marshalling signed TX to binary: %w", err)
	}

	return signedData, nil
}

// Signs an arbitrary message using the wallet's private key
func (w *LocalWallet) SignMessage(message []byte) ([]byte, error) {
	messageHash := accounts.TextHash(message)
	signedMessage, err := crypto.Sign(messageHash, w.nodePrivateKey)
	if err != nil {
		return nil, fmt.Errorf("Error signing message: %w", err)
	}

	// fix the ECDSA 'v' (see https://medium.com/mycrypto/the-magic-of-digital-signatures-on-ethereum-98fe184dc9c7#:~:text=The%20version%20number,2%E2%80%9D%20was%20introduced)
	signedMessage[crypto.RecoveryIDOffset] += 27
	return signedMessage, nil
}

// Initialize the wallet keystore from a mnemonic and derivation path
func (w *LocalWallet) initializeKeystore(derivationPath string, walletIndex uint, mnemonic string) error {
	// Get the wallet password
	password, hasPassword := w.passwordManager.Get()
	if !hasPassword {
		return fmt.Errorf("password has not been set yet")
	}

	// Generate the seed from the mnemonic
	w.seed = bip39.NewSeed(mnemonic, "")

	// Create the master key
	var err error
	w.masterKey, err = hdkeychain.NewMaster(w.seed, &chaincfg.MainNetParams)
	if err != nil {
		return fmt.Errorf("error creating wallet master key: %w", err)
	}

	// Encrypt the seed with the password
	encryptedSeed, err := w.encryptor.Encrypt(w.seed, string(password))
	if err != nil {
		return fmt.Errorf("error encrypting wallet seed: %w", err)
	}

	// Create a new wallet keystore
	keystore := &keystore.WalletKeystore{
		Crypto:         encryptedSeed,
		Name:           w.encryptor.Name(),
		Version:        w.encryptor.Version(),
		UUID:           uuid.New(),
		DerivationPath: derivationPath,
		WalletIndex:    walletIndex,
		NextAccount:    0,
	}

	// Save it
	w.keystoreManager.Set(keystore)
	err = w.keystoreManager.Save()
	if err != nil {
		return fmt.Errorf("error saving new wallet keystore: %w", err)
	}

	// Load the derived key and update the node address
	err = w.loadKeyFromKeystore(keystore, password, true)
	if err != nil {
		return fmt.Errorf("error loading wallet key from keystore: %w", err)
	}
	return nil
}

// Load the node wallet's private key from the keystore on disk
func (w *LocalWallet) loadKeyFromKeystore(keystore *keystore.WalletKeystore, password []byte, updateAddressFile bool) error {
	// Upgrade legacy wallets to include derivation paths
	if keystore.DerivationPath == "" {
		keystore.DerivationPath = DefaultNodeKeyPath
	}

	// Decrypt the seed
	var err error
	w.seed, err = w.encryptor.Decrypt(keystore.Crypto, string(password))
	if err != nil {
		return fmt.Errorf("error decrypting wallet keystore: %w", err)
	}

	// Create the master key
	w.masterKey, err = hdkeychain.NewMaster(w.seed, &chaincfg.MainNetParams)
	if err != nil {
		return fmt.Errorf("error creating wallet master key: %w", err)
	}

	// Get the derived key
	derivedKey, path, err := w.getDerivedKey(keystore.WalletIndex)
	if err != nil {
		return fmt.Errorf("error getting node wallet derived key: %w", err)
	}

	// Get the private key
	privateKey, err := derivedKey.ECPrivKey()
	if err != nil {
		return fmt.Errorf("error getting node wallet private key: %w", err)
	}
	privateKeyECDSA := privateKey.ToECDSA()

	// Store it
	w.nodePrivateKey = privateKeyECDSA
	w.nodeKeyPath = path

	// Make sure the pubkey matches the node address
	derivedAddress := crypto.PubkeyToAddress(w.nodePrivateKey.PublicKey)

	if updateAddressFile {
		// Set the address to the derived address and save it
		w.addressManager.Set(derivedAddress)
		err = w.addressManager.Save()
		if err != nil {
			return fmt.Errorf("error saving wallet address file for address derived from keystore (%s): %w", derivedAddress.Hex(), err)
		}
	}
	return nil
}

// Get the derived key & derivation path for the account at the index
func (w *LocalWallet) getDerivedKey(index uint) (*hdkeychain.ExtendedKey, string, error) {
	keystore, isSet := w.keystoreManager.Get()
	if !isSet {
		return nil, "", fmt.Errorf("wallet keystore has not been set")
	}

	// Get the derivation path
	if keystore.DerivationPath == "" {
		keystore.DerivationPath = DefaultNodeKeyPath
	}
	derivationPath := fmt.Sprintf(keystore.DerivationPath, index)

	// Parse derivation path
	path, err := accounts.ParseDerivationPath(derivationPath)
	if err != nil {
		return nil, "", fmt.Errorf("invalid node key derivation path '%s': %w", derivationPath, err)
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
			return w.getDerivedKey(index + 1)
		} else if err != nil {
			return nil, "", fmt.Errorf("invalid child key at depth %d: %w", i, err)
		}
	}

	// Return
	return key, derivationPath, nil
}
