package lodestar

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/rocket-pool/rocketpool-go/types"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
	eth2ks "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"

	"github.com/rocket-pool/smartnode/shared/services/passwords"
	keystore "github.com/rocket-pool/smartnode/shared/services/wallet/keystore"
	hexutil "github.com/rocket-pool/smartnode/shared/utils/hex"
)

// Config
const (
	KeystoreDir   = "lodestar"
	SecretsDir    = "secrets"
	ValidatorsDir = "validators"
	KeyFileName   = "voting-keystore.json"
	DirMode       = 0700
	FileMode      = 0600
)

// Lodestar keystore
type Keystore struct {
	keystorePath string
	pm           *passwords.PasswordManager
	encryptor    *eth2ks.Encryptor
}

// Encrypted validator key store
type validatorKey struct {
	Crypto  map[string]interface{} `json:"crypto"`
	Version uint                   `json:"version"`
	UUID    uuid.UUID              `json:"uuid"`
	Path    string                 `json:"path"`
	Pubkey  types.ValidatorPubkey  `json:"pubkey"`
}

// Create new lodestar keystore
func NewKeystore(keystorePath string, passwordManager *passwords.PasswordManager) *Keystore {
	return &Keystore{
		keystorePath: keystorePath,
		pm:           passwordManager,
		encryptor:    eth2ks.New(eth2ks.WithCipher("scrypt")),
	}
}

// Get the keystore directory
func (ks *Keystore) GetKeystoreDir() string {
	return filepath.Join(ks.keystorePath, KeystoreDir)
}

// Store a validator key
func (ks *Keystore) StoreValidatorKey(key *eth2types.BLSPrivateKey, derivationPath string) error {

	// Get validator pubkey
	pubkey := types.BytesToValidatorPubkey(key.PublicKey().Marshal())

	// Create a new password
	password, err := keystore.GenerateRandomPassword()
	if err != nil {
		return fmt.Errorf("Could not generate random password: %w", err)
	}

	// Encrypt key
	encryptedKey, err := ks.encryptor.Encrypt(key.Marshal(), password)
	if err != nil {
		return fmt.Errorf("Could not encrypt validator key: %w", err)
	}

	// Create key store
	keyStore := validatorKey{
		Crypto:  encryptedKey,
		Version: ks.encryptor.Version(),
		UUID:    uuid.New(),
		Path:    derivationPath,
		Pubkey:  pubkey,
	}

	// Encode key store
	keyStoreBytes, err := json.Marshal(keyStore)
	if err != nil {
		return fmt.Errorf("Could not encode validator key: %w", err)
	}

	// Get secret file path
	secretFilePath := filepath.Join(ks.keystorePath, KeystoreDir, SecretsDir, hexutil.AddPrefix(pubkey.Hex()))

	// Create secrets dir
	if err := os.MkdirAll(filepath.Dir(secretFilePath), DirMode); err != nil {
		return fmt.Errorf("Could not create validator secrets folder: %w", err)
	}

	// Write secret to disk
	if err := os.WriteFile(secretFilePath, []byte(password), FileMode); err != nil {
		return fmt.Errorf("Could not write validator secret to disk: %w", err)
	}

	// Get key file path
	keyFilePath := filepath.Join(ks.keystorePath, KeystoreDir, ValidatorsDir, hexutil.AddPrefix(pubkey.Hex()), KeyFileName)

	// Create key dir
	if err := os.MkdirAll(filepath.Dir(keyFilePath), DirMode); err != nil {
		return fmt.Errorf("Could not create validator key folder: %w", err)
	}

	// Write key store to disk
	if err := os.WriteFile(keyFilePath, keyStoreBytes, FileMode); err != nil {
		return fmt.Errorf("Could not write validator key to disk: %w", err)
	}

	// Return
	return nil

}

// Load a private key
func (ks *Keystore) LoadValidatorKey(pubkey types.ValidatorPubkey) (*eth2types.BLSPrivateKey, error) {

	// Get key file path
	keyFilePath := filepath.Join(ks.keystorePath, KeystoreDir, ValidatorsDir, hexutil.AddPrefix(pubkey.Hex()), KeyFileName)

	// Read the key file
	_, err := os.Stat(keyFilePath)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("couldn't open the Lodestar keystore for pubkey %s: %w", pubkey.Hex(), err)
	}
	bytes, err := os.ReadFile(keyFilePath)
	if err != nil {
		return nil, fmt.Errorf("couldn't read the Lodestar keystore for pubkey %s: %w", pubkey.Hex(), err)
	}

	// Unmarshal the keystore
	var keystore validatorKey
	err = json.Unmarshal(bytes, &keystore)
	if err != nil {
		return nil, fmt.Errorf("error deserializing Lodestar keystore for pubkey %s: %w", pubkey.Hex(), err)
	}

	// Get secret file path
	secretFilePath := filepath.Join(ks.keystorePath, KeystoreDir, SecretsDir, hexutil.AddPrefix(pubkey.Hex()))

	// Read secret from disk
	_, err = os.Stat(secretFilePath)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("couldn't open the Lodestar secret for pubkey %s: %w", pubkey.Hex(), err)
	}
	bytes, err = os.ReadFile(secretFilePath)
	if err != nil {
		return nil, fmt.Errorf("couldn't read the Lodestar secret for pubkey %s: %w", pubkey.Hex(), err)
	}

	// Decrypt key
	password := string(bytes)
	decryptedKey, err := ks.encryptor.Decrypt(keystore.Crypto, password)
	if err != nil {
		return nil, fmt.Errorf("couldn't decrypt keystore for pubkey %s: %w", pubkey.Hex(), err)
	}
	privateKey, err := eth2types.BLSPrivateKeyFromBytes(decryptedKey)
	if err != nil {
		return nil, fmt.Errorf("error recreating private key for validator %s: %w", keystore.Pubkey.Hex(), err)
	}

	// Verify the private key matches the public key
	reconstructedPubkey := types.BytesToValidatorPubkey(privateKey.PublicKey().Marshal())
	if reconstructedPubkey != pubkey {
		return nil, fmt.Errorf("private keystore file %s claims to be for validator %s but it's for validator %s", keyFilePath, pubkey.Hex(), reconstructedPubkey.Hex())
	}

	return privateKey, nil

}
