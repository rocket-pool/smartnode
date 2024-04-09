package validator

import (
	"bytes"
	"fmt"

	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/node/validator"
	"github.com/rocket-pool/node-manager-core/node/wallet"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/utils"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
	types "github.com/wealdtech/go-eth2-types/v2"
)

// Config
const (
	MaxValidatorKeyRecoverAttempts uint64 = 1000
)

// A validator private/public key pair
type ValidatorKey struct {
	PublicKey      beacon.ValidatorPubkey
	PrivateKey     *eth2types.BLSPrivateKey
	DerivationPath string
	WalletIndex    uint64
}

// Wallet management
type ValidatorManager struct {
	cfg             *config.SmartNodeConfig
	rp              *rocketpool.RocketPool
	wallet          *wallet.Wallet
	queryMgr        *eth.QueryManager
	keystoreManager *validator.ValidatorManager
	nextAccount     uint64
}

func NewValidatorManager(cfg *config.SmartNodeConfig, rp *rocketpool.RocketPool, walletImpl *wallet.Wallet, queryMgr *eth.QueryManager) (*ValidatorManager, error) {
	// Make a validator manager
	validatorManager := validator.NewValidatorManager(cfg.GetValidatorsFolderPath())

	// Make a new mgr
	mgr := &ValidatorManager{
		cfg:             cfg,
		rp:              rp,
		wallet:          walletImpl,
		queryMgr:        queryMgr,
		keystoreManager: validatorManager,
	}

	// Load the next account
	var err error
	mgr.nextAccount, err = loadNextAccount(cfg.GetNextAccountFilePath())
	if err != nil {
		return nil, err
	}

	return mgr, nil
}

// Get the number of validator keys recorded in the wallet
func (m *ValidatorManager) GetValidatorKeyCount() (uint64, error) {
	err := m.checkIfReady()
	if err != nil {
		return 0, err
	}
	return m.nextAccount, nil
}

// Get a validator key by index
func (m *ValidatorManager) GetValidatorKeyAt(index uint64) (*eth2types.BLSPrivateKey, error) {
	err := m.checkIfReady()
	if err != nil {
		return nil, err
	}

	// Return validator key
	key, _, err := m.getValidatorPrivateKey(index)
	return key, err
}

// Stores a validator key into all of the wallet's keystores
func (m *ValidatorManager) StoreValidatorKey(key *eth2types.BLSPrivateKey, path string) error {
	return m.keystoreManager.StoreKey(key, path)
}

// Loads a validator key from the wallet's keystores
func (m *ValidatorManager) LoadValidatorKey(pubkey beacon.ValidatorPubkey) (*eth2types.BLSPrivateKey, error) {
	return m.keystoreManager.LoadKey(pubkey)
}

// Returns the next validator key to generate, optionally saving it
func (m *ValidatorManager) GetNextValidatorKey(save bool) (*eth2types.BLSPrivateKey, uint64, error) {
	err := m.checkIfReady()
	if err != nil {
		return nil, 0, err
	}

	// Get account index
	index := m.nextAccount

	// Get validator key
	key, path, err := m.getValidatorPrivateKey(index)
	if err != nil {
		return nil, 0, err
	}

	if save {
		// Update keystores
		err = m.StoreValidatorKey(key, path)
		if err != nil {
			return nil, 0, err
		}

		// Increment the next account
		m.nextAccount++
		saveNextAccount(m.nextAccount, m.cfg.GetNextAccountFilePath())
	}

	// Return validator key
	return key, index, nil
}

// Recover a set of validator keys by their public key
func (m *ValidatorManager) GetValidatorKeys(startIndex uint64, length uint64) ([]ValidatorKey, error) {
	err := m.checkIfReady()
	if err != nil {
		return nil, err
	}

	validatorKeys := make([]ValidatorKey, 0, length)
	for index := startIndex; index < startIndex+length; index++ {
		key, path, err := m.getValidatorPrivateKey(index)
		if err != nil {
			return nil, fmt.Errorf("error getting validator key for index %d: %w", index, err)
		}
		validatorKey := ValidatorKey{
			PublicKey:      beacon.ValidatorPubkey(key.PublicKey().Marshal()),
			PrivateKey:     key,
			DerivationPath: path,
			WalletIndex:    index,
		}
		validatorKeys = append(validatorKeys, validatorKey)
	}

	return validatorKeys, nil
}

// Save a validator key
func (m *ValidatorManager) SaveValidatorKey(key ValidatorKey) error {
	// Update account index
	if key.WalletIndex >= m.nextAccount {
		m.nextAccount = key.WalletIndex + 1
	}

	// Update keystores
	err := m.keystoreManager.StoreKey(key.PrivateKey, key.DerivationPath)
	if err != nil {
		return fmt.Errorf("could not store validator %s key: %w", key.PublicKey.HexWithPrefix(), err)
	}
	saveNextAccount(m.nextAccount, m.cfg.GetNextAccountFilePath())

	// Return
	return nil
}

// Recover a validator key by public key
func (m *ValidatorManager) RecoverValidatorKey(pubkey beacon.ValidatorPubkey, startIndex uint64) (uint64, error) {
	err := m.checkIfReady()
	if err != nil {
		return 0, err
	}

	// Find matching validator key
	var index uint64
	var validatorKey *eth2types.BLSPrivateKey
	var derivationPath string
	for index = 0; index < MaxValidatorKeyRecoverAttempts; index++ {
		if key, path, err := m.getValidatorPrivateKey(index + startIndex); err != nil {
			return 0, err
		} else if bytes.Equal(pubkey[:], key.PublicKey().Marshal()) {
			validatorKey = key
			derivationPath = path
			break
		}
	}

	// Check validator key
	if validatorKey == nil {
		return 0, fmt.Errorf("validator %s key not found", pubkey.Hex())
	}

	// Update account index
	nextIndex := index + startIndex + 1
	if nextIndex > m.nextAccount {
		m.nextAccount = nextIndex
	}

	// Update keystores
	err = m.keystoreManager.StoreKey(validatorKey, derivationPath)
	if err != nil {
		return 0, fmt.Errorf("error storing validator %s key: %w", pubkey.HexWithPrefix(), err)
	}
	saveNextAccount(m.nextAccount, m.cfg.GetNextAccountFilePath())

	// Return
	return index + startIndex, nil
}

// Test recovery of a validator key by public key
func (m *ValidatorManager) TestRecoverValidatorKey(pubkey beacon.ValidatorPubkey, startIndex uint64) (uint64, error) {
	err := m.checkIfReady()
	if err != nil {
		return 0, err
	}

	// Find matching validator key
	var index uint64
	var validatorKey *eth2types.BLSPrivateKey
	for index = 0; index < MaxValidatorKeyRecoverAttempts; index++ {
		if key, _, err := m.getValidatorPrivateKey(index + startIndex); err != nil {
			return 0, err
		} else if bytes.Equal(pubkey[:], key.PublicKey().Marshal()) {
			validatorKey = key
			break
		}
	}

	// Check validator key
	if validatorKey == nil {
		return 0, fmt.Errorf("validator %s key not found", pubkey.Hex())
	}

	// Return
	return index + startIndex, nil
}

// Get a validator private key by index
func (m *ValidatorManager) getValidatorPrivateKey(index uint64) (*eth2types.BLSPrivateKey, string, error) {
	// Get derivation path
	derivationPath := fmt.Sprintf(ValidatorKeyPath, index)

	// Get private key
	privateKeyBytes, err := m.wallet.GenerateValidatorKey(derivationPath)
	if err != nil {
		return nil, "", fmt.Errorf("error getting validator %d private key: %w", index, err)
	}
	privateKey, err := types.BLSPrivateKeyFromBytes(privateKeyBytes)
	if err != nil {
		return nil, "", fmt.Errorf("error converting validator %d private key: %w", index, err)
	}
	return privateKey, derivationPath, nil
}

// Checks if the wallet is ready for validator key processing
func (m *ValidatorManager) checkIfReady() error {
	status, err := m.wallet.GetStatus()
	if err != nil {
		return err
	}
	return utils.CheckIfWalletReady(status)
}
