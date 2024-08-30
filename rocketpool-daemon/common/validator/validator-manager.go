package validator

import (
	"bytes"
	"fmt"
	"os"

	"github.com/goccy/go-json"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/node/validator"
	walletnode "github.com/rocket-pool/node-manager-core/node/wallet"
	walletcore "github.com/rocket-pool/node-manager-core/wallet"
	"github.com/rocket-pool/rocketpool-go/v2/minipool"
	"github.com/rocket-pool/rocketpool-go/v2/node"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/utils"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
	types "github.com/wealdtech/go-eth2-types/v2"
	eth2ks "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
	"gopkg.in/yaml.v3"
	"path/filepath"
)

// Config
const (
	MaxValidatorKeyRecoverAttempts uint64 = 1000
	bucketSize                     uint64 = 20
	bucketLimit                    uint64 = 2000
	pubkeyBatchSize                int    = 500
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
	wallet          *walletnode.Wallet
	queryMgr        *eth.QueryManager
	keystoreManager *validator.ValidatorManager
	nextAccount     uint64
	node            *node.Node
	minipoolManager *minipool.MinipoolManager
}

func NewValidatorManager(cfg *config.SmartNodeConfig, rp *rocketpool.RocketPool, walletImpl *walletnode.Wallet, queryMgr *eth.QueryManager) (*ValidatorManager, error) {
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
	err := mgr.initializeBindings()
	if err != nil {
		return nil, err
	}

	// Load the next account
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
		err = saveNextAccount(m.nextAccount, m.cfg.GetNextAccountFilePath())
		if err != nil {
			return nil, 0, err
		}
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
	err = saveNextAccount(m.nextAccount, m.cfg.GetNextAccountFilePath())
	if err != nil {
		return fmt.Errorf("could not store next validator account index: %w", err)
	}

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
	err = saveNextAccount(m.nextAccount, m.cfg.GetNextAccountFilePath())
	if err != nil {
		return 0, fmt.Errorf("error storing next validator account index: %w", err)
	}

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

func (m *ValidatorManager) GetMinipools() (map[beacon.ValidatorPubkey]bool, error) {
	err := m.initializeBindings()
	if err != nil {
		return nil, err
	}

	err = m.queryMgr.Query(nil, nil, m.node.ValidatingMinipoolCount)
	if err != nil {
		return nil, fmt.Errorf("error getting node's validating minipool count: %w", err)
	}

	addresses, err := m.node.GetValidatingMinipoolAddresses(m.node.ValidatingMinipoolCount.Formatted(), nil)
	if err != nil {
		return nil, fmt.Errorf("error getting node's validating minipool addresses: %w", err)
	}

	mps, err := m.minipoolManager.CreateMinipoolsFromAddresses(addresses, false, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating bindings for node's validating minipools: %w", err)
	}

	publicKeySet := map[beacon.ValidatorPubkey]bool{}
	zeroPublicKey := beacon.ValidatorPubkey{}

	err = m.queryMgr.BatchQuery(len(mps), pubkeyBatchSize, func(mc *batch.MultiCaller, i int) error {
		mps[i].Common().Pubkey.AddToQuery(mc)
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting node's validating minipool pubkeys: %w", err)
	}

	for _, mp := range mps {
		publicKey := mp.Common().Pubkey.Get()
		if publicKey != zeroPublicKey {
			publicKeySet[publicKey] = true
		}
	}

	return publicKeySet, nil
}

func (m *ValidatorManager) LoadCustomKeyPasswords() (map[string]string, error) {
	passwordFile := m.cfg.GetCustomKeyPasswordFilePath()
	fileBytes, err := os.ReadFile(passwordFile)
	if err != nil {
		return nil, fmt.Errorf("password file could not be loaded: %w", err)
	}
	passwords := map[string]string{}
	err = yaml.Unmarshal(fileBytes, &passwords)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling custom keystore password file: %w", err)
	}
	return passwords, nil
}

func (m *ValidatorManager) ReadCustomKeystore(file os.DirEntry) (api.ValidatorKeystore, error) {
	bytes, err := os.ReadFile(filepath.Join(m.cfg.GetCustomKeyPath(), file.Name()))
	if err != nil {
		return api.ValidatorKeystore{}, fmt.Errorf("error reading custom keystore %s: %w", file.Name(), err)
	}
	keystore := api.ValidatorKeystore{}
	err = json.Unmarshal(bytes, &keystore)
	if err != nil {
		return api.ValidatorKeystore{}, fmt.Errorf("error deserializing custom keystore %s: %w", file.Name(), err)
	}
	return keystore, nil
}

func (m *ValidatorManager) DecryptCustomKeystore(keystore api.ValidatorKeystore, password string) (eth2types.BLSPrivateKey, error) {
	kdf, exists := keystore.Crypto["kdf"]
	if !exists {
		return eth2types.BLSPrivateKey{}, fmt.Errorf("error processing custom keystore: \"crypto\" didn't contain a subkey named \"kdf\"")
	}
	kdfMap := kdf.(map[string]interface{})
	function, exists := kdfMap["function"]
	if !exists {
		return eth2types.BLSPrivateKey{}, fmt.Errorf("error processing custom keystore: \"crypto.kdf\" didn't contain a subkey named \"function\"")
	}
	functionString := function.(string)

	encryptor := eth2ks.New(eth2ks.WithCipher(functionString))
	decryptedKey, err := encryptor.Decrypt(keystore.Crypto, password)
	if err != nil {
		return eth2types.BLSPrivateKey{}, fmt.Errorf("error decrypting keystore: %w", err)
	}
	privateKey, err := eth2types.BLSPrivateKeyFromBytes(decryptedKey)
	if err != nil {
		return eth2types.BLSPrivateKey{}, fmt.Errorf("error recreating private key: %w", err)
	}
	return *privateKey, nil
}

func (m *ValidatorManager) LoadFiles() ([]os.DirEntry, error) {
	customKeyDir := m.cfg.GetCustomKeyPath()
	info, err := os.Stat(customKeyDir)

	if os.IsNotExist(err) || !info.IsDir() {
		err := fmt.Errorf("error loading custom keystore location: %w", err)
		return nil, err
	}

	keyFiles, err := os.ReadDir(customKeyDir)
	if err != nil {
		err := fmt.Errorf("error enumerating custom keystores: %w", err)
		return nil, err
	}

	if err := eth2types.InitBLS(); err != nil {
		err := fmt.Errorf("error initializing BLS: %w", err)
		return nil, err
	}
	return keyFiles, nil
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

func (m *ValidatorManager) initializeBindings() error {
	status, err := m.getWalletStatus()
	if err != nil {
		return err
	}

	rpNode, err := node.NewNode(m.rp, status.Wallet.WalletAddress)
	if err != nil {
		return err
	}

	mpMgr, err := minipool.NewMinipoolManager(m.rp)
	if err != nil {
		return err
	}

	m.node = rpNode
	m.minipoolManager = mpMgr

	return nil
}

func (m *ValidatorManager) getWalletStatus() (*walletcore.WalletStatus, error) {
	status, err := m.wallet.GetStatus()
	if err != nil {
		return &status, err
	}
	if !walletcore.IsWalletReady(status) {
		return &status, fmt.Errorf("wallet is not ready")
	}
	return &status, nil
}
