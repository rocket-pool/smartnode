package prysm

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/rocketpool-go/types"
	rpkeystore "github.com/rocket-pool/smartnode/rocketpool-daemon/common/wallet/keystore"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
	eth2ks "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
)

// Config
const (
	KeystoreDir              string      = "prysm-non-hd"
	WalletDir                string      = "direct"
	AccountsDir              string      = "accounts"
	KeystoreFileName         string      = "all-accounts.keystore.json"
	ConfigFileName           string      = "keymanageropts.json"
	KeystorePasswordFileName string      = "secret"
	DirMode                  fs.FileMode = 0770
	FileMode                 fs.FileMode = 0640

	DirectEIPVersion = "EIP-2335"
)

// Prysm keystore
type Keystore struct {
	keystorePath string
	as           *accountStore
	encryptor    *eth2ks.Encryptor
}

// Encrypted validator keystore
type validatorKeystore struct {
	Crypto  map[string]interface{} `json:"crypto"`
	Name    string                 `json:"name"`
	Version uint                   `json:"version"`
	UUID    uuid.UUID              `json:"uuid"`
	Pubkey  string                 `json:"pubkey"`
}
type accountStore struct {
	PrivateKeys [][]byte `json:"private_keys"`
	PublicKeys  [][]byte `json:"public_keys"`
}

// Prysm direct wallet config
type walletConfig struct {
	DirectEIPVersion string `json:"direct_eip_version"`
}

// Create new prysm keystore
func NewKeystore(keystorePath string) *Keystore {
	return &Keystore{
		keystorePath: keystorePath,
		encryptor:    eth2ks.New(),
	}
}

// Get the keystore directory
func (ks *Keystore) GetKeystoreDir() string {
	return filepath.Join(ks.keystorePath, KeystoreDir)
}

// Store a validator key
func (ks *Keystore) StoreValidatorKey(key *eth2types.BLSPrivateKey, derivationPath string) error {

	// Initialize the account store
	if err := ks.initialize(); err != nil {
		return err
	}

	// Cancel if validator key already exists in account store
	for ki := 0; ki < len(ks.as.PrivateKeys); ki++ {
		if bytes.Equal(key.Marshal(), ks.as.PrivateKeys[ki]) || bytes.Equal(key.PublicKey().Marshal(), ks.as.PublicKeys[ki]) {
			return nil
		}
	}

	// Add validator key to account store
	ks.as.PrivateKeys = append(ks.as.PrivateKeys, key.Marshal())
	ks.as.PublicKeys = append(ks.as.PublicKeys, key.PublicKey().Marshal())

	// Encode account store
	asBytes, err := json.Marshal(ks.as)
	if err != nil {
		return fmt.Errorf("Could not encode validator account store: %w", err)
	}

	// Get the keystore account password
	passwordFilePath := filepath.Join(ks.keystorePath, KeystoreDir, WalletDir, AccountsDir, KeystorePasswordFileName)
	passwordBytes, err := os.ReadFile(passwordFilePath)
	if err != nil {
		return fmt.Errorf("Error reading account password file: %w", err)
	}
	password := string(passwordBytes)

	// Encrypt account store
	asEncrypted, err := ks.encryptor.Encrypt(asBytes, password)
	if err != nil {
		return fmt.Errorf("Could not encrypt validator account store: %w", err)
	}

	// Create new keystore
	keystore := validatorKeystore{
		Crypto:  asEncrypted,
		Name:    ks.encryptor.Name(),
		Version: ks.encryptor.Version(),
		UUID:    uuid.New(),
	}

	// Encode key store
	ksBytes, err := json.Marshal(keystore)
	if err != nil {
		return fmt.Errorf("Could not encode validator keystore: %w", err)
	}

	// Get file paths
	keystoreFilePath := filepath.Join(ks.keystorePath, KeystoreDir, WalletDir, AccountsDir, KeystoreFileName)
	configFilePath := filepath.Join(ks.keystorePath, KeystoreDir, WalletDir, ConfigFileName)

	// Create keystore dir
	if err := os.MkdirAll(filepath.Dir(keystoreFilePath), DirMode); err != nil {
		return fmt.Errorf("Could not create keystore folder: %w", err)
	}

	// Write keystore to disk
	if err := os.WriteFile(keystoreFilePath, ksBytes, FileMode); err != nil {
		return fmt.Errorf("Could not write keystore to disk: %w", err)
	}

	// Return if wallet config file exists
	if _, err := os.Stat(configFilePath); !os.IsNotExist(err) {
		return nil
	}

	// Create & encode wallet config
	configBytes, err := json.Marshal(walletConfig{
		DirectEIPVersion: DirectEIPVersion,
	})
	if err != nil {
		return fmt.Errorf("Could not encode wallet config: %w", err)
	}

	// Write wallet config to disk
	if err := os.WriteFile(configFilePath, configBytes, FileMode); err != nil {
		return fmt.Errorf("Could not write wallet config to disk: %w", err)
	}

	// Return
	return nil

}

// Initialize the account store
func (ks *Keystore) initialize() error {

	// Cancel if already initialized
	if ks.as != nil {
		return nil
	}

	// Create the random keystore password if it doesn't exist
	var password string
	passwordFilePath := filepath.Join(ks.keystorePath, KeystoreDir, WalletDir, AccountsDir, KeystorePasswordFileName)
	_, err := os.Stat(passwordFilePath)
	if os.IsNotExist(err) {
		// Create a new password
		password, err = rpkeystore.GenerateRandomPassword()
		if err != nil {
			return fmt.Errorf("Could not generate random password: %w", err)
		}

		// Encode it
		passwordBytes := []byte(password)

		// Write it
		err := os.MkdirAll(filepath.Dir(passwordFilePath), DirMode)
		if err != nil {
			return fmt.Errorf("Error creating account password directory: %w", err)
		}
		err = os.WriteFile(passwordFilePath, passwordBytes, FileMode)
		if err != nil {
			return fmt.Errorf("Error writing account password file: %w", err)
		}
	}

	// Get the random keystore password
	passwordBytes, err := os.ReadFile(passwordFilePath)
	if err != nil {
		return fmt.Errorf("Error opening account password file: %w", err)
	}
	password = string(passwordBytes)

	// Read keystore file; initialize empty account store if it doesn't exist
	ksBytes, err := os.ReadFile(filepath.Join(ks.keystorePath, KeystoreDir, WalletDir, AccountsDir, KeystoreFileName))
	if err != nil {
		ks.as = &accountStore{}
		return nil
	}

	// Decode keystore
	keystore := &validatorKeystore{}
	if err = json.Unmarshal(ksBytes, keystore); err != nil {
		return fmt.Errorf("Could not decode validator keystore: %w", err)
	}

	// Decrypt account store
	asBytes, err := ks.encryptor.Decrypt(keystore.Crypto, password)
	if err != nil {
		return fmt.Errorf("Could not decrypt validator account store: %w", err)
	}

	// Decode account store
	as := &accountStore{}
	if err = json.Unmarshal(asBytes, as); err != nil {
		return fmt.Errorf("Could not decode validator account store: %w", err)
	}
	if len(as.PrivateKeys) != len(as.PublicKeys) {
		return errors.New("Validator account store private and public key counts do not match")
	}

	// Set account store & return
	ks.as = as
	return nil

}

// Load a private key
func (ks *Keystore) LoadValidatorKey(pubkey beacon.ValidatorPubkey) (*eth2types.BLSPrivateKey, error) {

	// Initialize the account store
	err := ks.initialize()
	if err != nil {
		return nil, err
	}

	// Find the validator key in the account store
	for ki := 0; ki < len(ks.as.PrivateKeys); ki++ {
		if bytes.Equal(pubkey.Bytes(), ks.as.PublicKeys[ki]) {
			decryptedKey := ks.as.PrivateKeys[ki]
			privateKey, err := eth2types.BLSPrivateKeyFromBytes(decryptedKey)
			if err != nil {
				return nil, fmt.Errorf("error recreating private key for validator %s: %w", pubkey.Hex(), err)
			}

			// Verify the private key matches the public key
			reconstructedPubkey := types.BytesToValidatorPubkey(privateKey.PublicKey().Marshal())
			if reconstructedPubkey != pubkey {
				return nil, fmt.Errorf("Prysm's keystore has a key that claims to be for validator %s but it's for validator %s", pubkey.Hex(), reconstructedPubkey.Hex())
			}

			return privateKey, nil
		}
	}

	// Return nothing if the private key wasn't found
	return nil, nil

}
