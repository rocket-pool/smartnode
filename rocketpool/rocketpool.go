package rocketpool

import (
    "fmt"
    "sync"
    "time"

    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/ethclient"

    "github.com/rocket-pool/rocketpool-go/contracts"
)


// Cache settings
const CACHE_TTL = 60 // Seconds


// Cached data types
type cachedAddress struct {
    address *common.Address
    time int64
}
type cachedABI struct {
    abi *abi.ABI
    time int64
}


// Rocket Pool contract manager
type RocketPool struct {
    Client          *ethclient.Client
    RocketStorage   *contracts.RocketStorage
    addresses       map[string]cachedAddress
    abis            map[string]cachedABI
    addressesLock   sync.RWMutex
    abisLock        sync.RWMutex
}


// Create new contract manager
func NewRocketPool(client *ethclient.Client, rocketStorageAddress common.Address) (*RocketPool, error) {

    // Initialize RocketStorage contract
    rocketStorage, err := contracts.NewRocketStorage(rocketStorageAddress, client)
    if err != nil {
        return nil, fmt.Errorf("Could not initialize Rocket Pool storage contract: %w", err)
    }

    // Create and return
    return &RocketPool{
        Client: client,
        RocketStorage: rocketStorage,
        addresses: make(map[string]cachedAddress),
        abis: make(map[string]cachedABI),
    }, nil

}


// Load a Rocket Pool contract address
func (rp *RocketPool) Address(contractName string) (*common.Address, error) {

    // Check for cached address
    if cached, ok := rp.getCachedAddress(contractName); ok {
        if (time.Now().Unix() - cached.time <= CACHE_TTL) {
            return cached.address, nil
        } else {
            rp.deleteCachedAddress(contractName)
        }
    }

    // Get address
    address, err := rp.RocketStorage.GetAddress(nil, crypto.Keccak256Hash([]byte("contract.name"), []byte(contractName)))
    if err != nil {
        return nil, fmt.Errorf("Could not get contract address: %w", err)
    }

    // Cache address
    rp.setCachedAddress(contractName, cachedAddress{
        address: &address,
        time: time.Now().Unix(),
    })

    // Return
    return &address, nil

}


// Load a Rocket Pool contract ABI
func (rp *RocketPool) ABI(contractName string) (*abi.ABI, error) {
    return nil, nil
}


// Load a Rocket Pool contract
func (rp *RocketPool) Get(contractName string) (*bind.BoundContract, error) {
    return nil, nil
}


// Create a Rocket Pool contract instance
func (rp *RocketPool) Make(contractName string, address common.Address) (*bind.BoundContract, error) {
    return nil, nil
}


// Address cache control
func (rp *RocketPool) getCachedAddress(contractName string) (cachedAddress, bool) {
    rp.addressesLock.RLock()
    defer rp.addressesLock.RUnlock()
    value, ok := rp.addresses[contractName]
    return value, ok
}
func (rp *RocketPool) setCachedAddress(contractName string, value cachedAddress) {
    rp.addressesLock.Lock()
    defer rp.addressesLock.Unlock()
    rp.addresses[contractName] = value
}
func (rp *RocketPool) deleteCachedAddress(contractName string) {
    rp.addressesLock.Lock()
    defer rp.addressesLock.Unlock()
    delete(rp.addresses, contractName)
}

