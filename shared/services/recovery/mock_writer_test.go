package recovery

import (
	"sync"

	eth2types "github.com/wealdtech/go-eth2-types/v2"

	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
)

// StoreCall records a call to StoreValidatorKey.
type StoreCall struct {
	Key  *eth2types.BLSPrivateKey
	Path string
}

// MockKeyWriter is a thread-safe spy implementation of KeyWriter that records all write attempts.
type MockKeyWriter struct {
	mu                     sync.Mutex
	saveValidatorKeyCalls  []wallet.ValidatorKey
	storeValidatorKeyCalls []StoreCall

	SaveError  error
	StoreError error
}

func NewMockKeyWriter() *MockKeyWriter {
	return &MockKeyWriter{
		saveValidatorKeyCalls:  make([]wallet.ValidatorKey, 0),
		storeValidatorKeyCalls: make([]StoreCall, 0),
	}
}

func (m *MockKeyWriter) SaveValidatorKey(key wallet.ValidatorKey) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.SaveError != nil {
		return m.SaveError
	}
	m.saveValidatorKeyCalls = append(m.saveValidatorKeyCalls, key)
	return nil
}

func (m *MockKeyWriter) StoreValidatorKey(key *eth2types.BLSPrivateKey, path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.StoreError != nil {
		return m.StoreError
	}
	m.storeValidatorKeyCalls = append(m.storeValidatorKeyCalls, StoreCall{Key: key, Path: path})
	return nil
}

func (m *MockKeyWriter) TotalWrites() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.saveValidatorKeyCalls) + len(m.storeValidatorKeyCalls)
}

func (m *MockKeyWriter) SaveCalls() []wallet.ValidatorKey {
	m.mu.Lock()
	defer m.mu.Unlock()
	copied := make([]wallet.ValidatorKey, len(m.saveValidatorKeyCalls))
	copy(copied, m.saveValidatorKeyCalls)
	return copied
}

func (m *MockKeyWriter) StoreCalls() []StoreCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	copied := make([]StoreCall, len(m.storeValidatorKeyCalls))
	copy(copied, m.storeValidatorKeyCalls)
	return copied
}

func (m *MockKeyWriter) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.saveValidatorKeyCalls = m.saveValidatorKeyCalls[:0]
	m.storeValidatorKeyCalls = m.storeValidatorKeyCalls[:0]
	m.SaveError = nil
	m.StoreError = nil
}

// MockKeyDeriver provides in-memory mock validator keys for derivation testing.
type MockKeyDeriver struct {
	mu          sync.RWMutex
	keysByIndex map[uint]wallet.ValidatorKey
}

func NewMockKeyDeriver() *MockKeyDeriver {
	return &MockKeyDeriver{
		keysByIndex: make(map[uint]wallet.ValidatorKey),
	}
}

func (m *MockKeyDeriver) AddKey(index uint, key wallet.ValidatorKey) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.keysByIndex[index] = key
}

func (m *MockKeyDeriver) GetValidatorKeys(startIndex uint, length uint) ([]wallet.ValidatorKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	res := make([]wallet.ValidatorKey, 0, length)
	for i := startIndex; i < startIndex+length; i++ {
		if k, ok := m.keysByIndex[i]; ok {
			res = append(res, k)
		}
	}
	return res, nil
}

// MockCustomKeyProvider provides custom keys for testing.
type MockCustomKeyProvider struct {
	mu   sync.RWMutex
	keys []CustomKey
	err  error
}

func NewMockCustomKeyProvider(keys []CustomKey, err error) *MockCustomKeyProvider {
	return &MockCustomKeyProvider{keys: keys, err: err}
}

func (m *MockCustomKeyProvider) GetCustomKeys() ([]CustomKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	return m.keys, nil
}

// MockInstalledKeyChecker checks installed keys for testing.
type MockInstalledKeyChecker struct {
	mu        sync.RWMutex
	installed map[types.ValidatorPubkey]bool
}

func NewMockInstalledKeyChecker(installedPubkeys ...types.ValidatorPubkey) *MockInstalledKeyChecker {
	m := &MockInstalledKeyChecker{installed: make(map[types.ValidatorPubkey]bool, len(installedPubkeys))}
	for _, pk := range installedPubkeys {
		m.installed[pk] = true
	}
	return m
}

func (m *MockInstalledKeyChecker) IsKeyInstalled(pubkey types.ValidatorPubkey) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.installed[pubkey], nil
}
