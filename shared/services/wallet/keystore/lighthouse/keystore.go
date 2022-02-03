package lighthouse

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
	eth2ks "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"

	"github.com/rocket-pool/smartnode/shared/services/passwords"
	keystore "github.com/rocket-pool/smartnode/shared/services/wallet/keystore"
	hexutil "github.com/rocket-pool/smartnode/shared/utils/hex"
)

// Config
const (
	KeystoreDir   = "lighthouse"
	SecretsDir    = "secrets"
	ValidatorsDir = "validators"
	KeyFileName   = "voting-keystore.json"
	DirMode       = 0700
	FileMode      = 0600
)

// Lighthouse keystore
type Keystore struct {
	keystorePath string
	pm           *passwords.PasswordManager
	encryptor    *eth2ks.Encryptor
}

// Encrypted validator key store
type validatorKey struct {
	Crypto  map[string]interface{}  `json:"crypto"`
	Version uint                    `json:"version"`
	UUID    uuid.UUID               `json:"uuid"`
	Path    string                  `json:"path"`
	Pubkey  rptypes.ValidatorPubkey `json:"pubkey"`
}

// Create new lighthouse keystore
func NewKeystore(keystorePath string, passwordManager *passwords.PasswordManager) *Keystore {
	return &Keystore{
		keystorePath: keystorePath,
		pm:           passwordManager,
		encryptor:    eth2ks.New(eth2ks.WithCipher("scrypt")),
	}
}

// Store a validator key
func (ks *Keystore) StoreValidatorKey(key *eth2types.BLSPrivateKey, derivationPath string) error {

	// Get validator pubkey
	pubkey := rptypes.BytesToValidatorPubkey(key.PublicKey().Marshal())

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
	if err := ioutil.WriteFile(secretFilePath, []byte(password), FileMode); err != nil {
		return fmt.Errorf("Could not write validator secret to disk: %w", err)
	}

	// Get key file path
	keyFilePath := filepath.Join(ks.keystorePath, KeystoreDir, ValidatorsDir, hexutil.AddPrefix(pubkey.Hex()), KeyFileName)

	// Create key dir
	if err := os.MkdirAll(filepath.Dir(keyFilePath), DirMode); err != nil {
		return fmt.Errorf("Could not create validator key folder: %w", err)
	}

	// Write key store to disk
	if err := ioutil.WriteFile(keyFilePath, keyStoreBytes, FileMode); err != nil {
		return fmt.Errorf("Could not write validator key to disk: %w", err)
	}

	// Return
	return nil

}
